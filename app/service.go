package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// ChatMessage is the UI-friendly form of a decrypted message.
type ChatMessage struct {
	SenderID   string
	ReceiverID string
	Content    string
	IsHeart    bool
	Incoming   bool
	CreatedAt  time.Time
}

// Service coordinates auth, persistence, websocket chat, and configuration.
type Service struct {
	mu           sync.RWMutex
	config       ClientConfig
	storage      *LocalStorage
	api          *APIClient
	currentUser  User
	localKeys    LocalKeys
	activeFriend Friend
	wsClient     *WSClient
	chatCancel   context.CancelFunc
	chatMessages chan ChatMessage
	chatStatus   chan string
}

// NewService constructs the client service and loads local preferences.
func NewService() (*Service, error) {
	config := LoadClientConfig()
	storage, err := NewLocalStorage(config.DataDir)
	if err != nil {
		return nil, err
	}

	if prefs, err := storage.LoadPreferences(); err == nil && strings.TrimSpace(prefs.ServerURL) != "" {
		config.ServerURL = strings.TrimSpace(prefs.ServerURL)
	}

	service := &Service{
		config:  config,
		storage: storage,
		api:     NewAPIClient(config.ServerURL),
	}

	if _, err := service.resumeSession(); err != nil {
		return nil, err
	}
	return service, nil
}

// Bootstrap attempts to restore the saved session.
func (s *Service) Bootstrap() (bool, error) {
	return s.resumeSession()
}

// Config returns the effective client configuration.
func (s *Service) Config() ClientConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// CurrentUser returns the authenticated user, if any.
func (s *Service) CurrentUser() (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentUser.ID == "" {
		return User{}, false
	}
	return s.currentUser, true
}

// UpdateServerURL persists a new server URL and updates the API client.
func (s *Service) UpdateServerURL(serverURL string) error {
	serverURL = strings.TrimSpace(serverURL)
	if serverURL == "" {
		return fmt.Errorf("server URL cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.ServerURL = serverURL
	s.api.baseURL = serverURL
	return s.storage.SavePreferences(Preferences{ServerURL: serverURL})
}

// Register creates a new account and provisions local keys.
func (s *Service) Register(username, email, password string) (User, error) {
	response, err := s.api.Register(username, email, password)
	if err != nil {
		return User{}, err
	}

	if err := s.ensureLocalKeys(response.UserID, true); err != nil {
		return User{}, err
	}

	if err := s.storage.SaveSession(response.UserID, response.Token); err != nil {
		return User{}, err
	}

	s.mu.Lock()
	s.currentUser = response.User
	s.mu.Unlock()
	return response.User, nil
}

// Login authenticates an account and restores or creates local keys.
func (s *Service) Login(username, password string) (User, error) {
	response, err := s.api.Login(username, password)
	if err != nil {
		return User{}, err
	}

	if err := s.ensureLocalKeys(response.UserID, false); err != nil {
		return User{}, err
	}

	if err := s.storage.SaveSession(response.UserID, response.Token); err != nil {
		return User{}, err
	}

	s.mu.Lock()
	s.currentUser = response.User
	s.mu.Unlock()
	return response.User, nil
}

// Logout clears the local session and stops any active chat connection.
func (s *Service) Logout() error {
	s.StopChat()
	if err := s.storage.ClearSession(); err != nil && !os.IsNotExist(err) {
		return err
	}

	s.api = NewAPIClient(s.config.ServerURL)
	s.mu.Lock()
	s.currentUser = User{}
	s.localKeys = LocalKeys{}
	s.activeFriend = Friend{}
	s.mu.Unlock()
	return nil
}

// Friends returns the current user's friend list.
func (s *Service) Friends() ([]Friend, error) {
	return s.api.GetFriends()
}

// FriendRequests returns the current user's pending requests.
func (s *Service) FriendRequests() ([]FriendRequest, error) {
	return s.api.GetFriendRequests()
}

// AddFriend sends a friend request by username.
func (s *Service) AddFriend(username string) error {
	return s.api.AddFriend(username)
}

// AcceptFriendRequest accepts a request and exchanges public keys.
func (s *Service) AcceptFriendRequest(request FriendRequest) error {
	if err := s.api.AcceptFriendRequest(request.ID); err != nil {
		return err
	}

	pubKey, _, err := GenerateKeyPair()
	if err != nil {
		return err
	}

	if err := s.api.UpdatePublicKey(request.SenderID, pubKey); err != nil {
		return err
	}

	return s.storage.SaveFriendPublicKey(request.SenderID, pubKey)
}

// BlockedUsers returns the local block list.
func (s *Service) BlockedUsers() ([]BlockedUser, error) {
	return s.api.GetBlockedUsers()
}

// LoadHistory returns decrypted chat history for a friend.
func (s *Service) LoadHistory(friend Friend, limit int) ([]ChatMessage, error) {
	s.mu.RLock()
	currentUser := s.currentUser
	s.mu.RUnlock()

	messages, err := s.api.GetMessageHistory(friend.FriendUserID, limit, 0)
	if err != nil {
		return nil, err
	}

	history := make([]ChatMessage, 0, len(messages))
	for _, message := range messages {
		history = append(history, ChatMessage{
			SenderID:   message.SenderID,
			ReceiverID: message.ReceiverID,
			Content:    message.Content,
			IsHeart:    message.IsHeart,
			Incoming:   message.SenderID != currentUser.ID,
			CreatedAt:  message.CreatedAt,
		})
	}

	return history, nil
}

// StartChat connects to the websocket and streams decrypted messages.
func (s *Service) StartChat(friend Friend) (<-chan ChatMessage, <-chan string, error) {
	s.StopChat()

	s.mu.RLock()
	privateKey := s.localKeys.PrivateKey
	token := s.api.token
	serverURL := s.config.ServerURL
	s.mu.RUnlock()

	if strings.TrimSpace(privateKey) == "" {
		return nil, nil, fmt.Errorf("local keys are missing")
	}

	ctx, cancel := context.WithCancel(context.Background())
	messages := make(chan ChatMessage, 64)
	status := make(chan string, 32)

	s.mu.Lock()
	s.chatCancel = cancel
	s.chatMessages = messages
	s.chatStatus = status
	s.activeFriend = friend
	s.mu.Unlock()

	go s.chatLoop(ctx, serverURL, token, friend, privateKey, messages, status)
	return messages, status, nil
}

// StopChat tears down the websocket connection and background goroutines.
func (s *Service) StopChat() {
	s.mu.Lock()
	cancel := s.chatCancel
	wsClient := s.wsClient
	s.chatCancel = nil
	s.chatMessages = nil
	s.chatStatus = nil
	s.activeFriend = Friend{}
	s.wsClient = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if wsClient != nil {
		wsClient.Disconnect()
	}
}

// SendMessage encrypts and sends a message to the active friend.
func (s *Service) SendMessage(content string, isHeart bool) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("message cannot be empty")
	}

	s.mu.RLock()
	friend := s.activeFriend
	wsClient := s.wsClient
	localKeys := s.localKeys
	s.mu.RUnlock()

	if wsClient == nil || !wsClient.IsConnected() {
		return fmt.Errorf("chat is not connected")
	}

	encrypted, err := EncryptMessage(friend.PublicKey, content)
	if err != nil {
		return err
	}

	if strings.TrimSpace(localKeys.PrivateKey) == "" {
		return fmt.Errorf("local private key is missing")
	}

	wsClient.SendMessage(friend.FriendUserID, encrypted, isHeart)
	return nil
}

func (s *Service) resumeSession() (bool, error) {
	_, token, err := s.storage.LoadSession()
	if err != nil {
		return false, nil
	}

	s.api.SetToken(token)
	user, err := s.api.GetMe()
	if err != nil {
		return false, nil
	}

	keys, err := s.storage.LoadKeys()
	if err != nil || keys.UserID == "" {
		if err := s.ensureLocalKeys(user.ID, false); err != nil {
			return false, err
		}
	} else {
		s.mu.Lock()
		s.localKeys = keys
		s.mu.Unlock()
	}

	s.mu.Lock()
	s.currentUser = user
	s.mu.Unlock()
	return true, nil
}

func (s *Service) ensureLocalKeys(userID string, forceGenerate bool) error {
	if !forceGenerate {
		if keys, err := s.storage.LoadKeys(); err == nil && keys.UserID != "" {
			s.mu.Lock()
			s.localKeys = keys
			s.mu.Unlock()
			return nil
		}
	}

	pubKey, privKey, err := GenerateKeyPair()
	if err != nil {
		return err
	}

	if err := s.storage.SaveKeys(userID, pubKey, privKey); err != nil {
		return err
	}

	s.mu.Lock()
	s.localKeys = LocalKeys{UserID: userID, PublicKey: pubKey, PrivateKey: privKey}
	s.mu.Unlock()
	return nil
}

func (s *Service) chatLoop(ctx context.Context, serverURL, token string, friend Friend, privateKey string, messages chan<- ChatMessage, status chan<- string) {
	defer close(messages)
	defer close(status)

	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

			signalStatus(status, "Connecting...")
		wsClient := NewWSClient(wsHostFromServerURL(serverURL), token)
		if err := wsClient.Connect(); err != nil {
			signalStatus(status, fmt.Sprintf("Reconnect failed: %v", err))
			if !sleepOrDone(ctx, backoff) {
				return
			}
			if backoff < 5*time.Second {
				backoff *= 2
			}
			continue
		}

		s.mu.Lock()
		s.wsClient = wsClient
		s.mu.Unlock()
		signalStatus(status, "Connected")
		backoff = time.Second

		connectionState := wsClient.ConnectionState()
		receive := wsClient.ReceiveMessages()

		for {
			select {
			case <-ctx.Done():
				wsClient.Disconnect()
				return
			case connected := <-connectionState:
				if !connected {
					signalStatus(status, "Disconnected, reconnecting...")
					goto reconnect
				}
			case msg := <-receive:
				if msg.SenderID != friend.FriendUserID {
					continue
				}

				decrypted, err := DecryptMessage(privateKey, msg.Content)
				if err != nil {
					signalStatus(status, fmt.Sprintf("Decrypt failed: %v", err))
					continue
				}

				message := ChatMessage{
					SenderID:   msg.SenderID,
					ReceiverID: msg.ReceiverID,
					Content:    decrypted,
					IsHeart:    msg.Type == "heart",
					Incoming:   true,
				}

				select {
				case messages <- message:
				default:
				}
			}
		}

	reconnect:
		wsClient.Disconnect()
		if !sleepOrDone(ctx, backoff) {
			return
		}
		if backoff < 5*time.Second {
			backoff *= 2
		}
	}
}

func signalStatus(status chan<- string, value string) {
	select {
	case status <- value:
	default:
	}
}

func sleepOrDone(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func wsHostFromServerURL(serverURL string) string {
	parsed, err := url.Parse(serverURL)
	if err != nil || parsed.Host == "" {
		return strings.TrimPrefix(strings.TrimPrefix(serverURL, "http://"), "https://")
	}
	return parsed.Host
}
