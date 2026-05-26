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
	// event listener
	eventClient   *WSClient
	eventCancel   context.CancelFunc
	eventMu       sync.Mutex
	eventHandlers []func()
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
	// start background event listener
	go s.startEventListener()
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
	// start background event listener
	go s.startEventListener()
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
	s.mu.RLock()
	pub := s.localKeys.PublicKey
	s.mu.RUnlock()
	return s.api.AddFriend(username, pub)
}

// AcceptFriendRequest accepts a request and exchanges public keys.
func (s *Service) AcceptFriendRequest(request FriendRequest) error {
	s.mu.RLock()
	myPublicKey := s.localKeys.PublicKey
	s.mu.RUnlock()

	if err := s.api.AcceptFriendRequest(request.ID, myPublicKey); err != nil {
		return err
	}

	// Fetch updated friend list to get the sender's public key
	friends, err := s.api.GetFriends()
	if err != nil {
		return err
	}

	// Find and save the new friend's public key
	for _, friend := range friends {
		if friend.FriendUserID == request.SenderID && friend.PublicKey != "" {
			return s.storage.SaveFriendPublicKey(request.SenderID, friend.PublicKey)
		}
	}

	return nil
}

// BlockedUsers returns the local block list.
func (s *Service) BlockedUsers() ([]BlockedUser, error) {
	return s.api.GetBlockedUsers()
}

// LoadHistory returns decrypted chat history for a friend.
func (s *Service) LoadHistory(friend Friend, limit int) ([]ChatMessage, error) {
	s.mu.RLock()
	currentUser := s.currentUser
	privateKey := s.localKeys.PrivateKey
	s.mu.RUnlock()

	messages, err := s.api.GetMessageHistory(friend.FriendUserID, limit, 0)
	if err != nil {
		return nil, err
	}

	history := make([]ChatMessage, 0, len(messages))
	for _, message := range messages {
		content := message.Content
		// Try to decrypt if we have a private key
		if strings.TrimSpace(privateKey) != "" && message.SenderID != currentUser.ID {
			if decrypted, err := DecryptMessage(privateKey, message.Content); err == nil {
				content = decrypted
			}
		}
		history = append(history, ChatMessage{
			SenderID:   message.SenderID,
			ReceiverID: message.ReceiverID,
			Content:    content,
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

	// If public key is missing, try to fetch fresh friend data from API
	if strings.TrimSpace(friend.PublicKey) == "" {
		updatedFriends, err := s.api.GetFriends()
		if err == nil {
			for _, f := range updatedFriends {
				if f.FriendUserID == friend.FriendUserID && f.PublicKey != "" {
					friend.PublicKey = f.PublicKey
					break
				}
			}
		}
	}

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

	// If public key is missing, try to fetch fresh friend data
	publicKey := friend.PublicKey
	if strings.TrimSpace(publicKey) == "" {
		updatedFriends, err := s.api.GetFriends()
		if err == nil {
			for _, f := range updatedFriends {
				if f.FriendUserID == friend.FriendUserID && f.PublicKey != "" {
					publicKey = f.PublicKey
					break
				}
			}
		}
	}

	if strings.TrimSpace(publicKey) == "" {
		return fmt.Errorf("friend's public key not available yet - wait for key exchange to complete")
	}

	encrypted, err := EncryptMessage(publicKey, content)
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
	// start background event listener
	go s.startEventListener()
	return true, nil
}

// RegisterEventHandler registers a callback invoked on friend-related events
func (s *Service) RegisterEventHandler(fn func()) {
	s.eventMu.Lock()
	defer s.eventMu.Unlock()
	s.eventHandlers = append(s.eventHandlers, fn)
}

func (s *Service) callEventHandlers() {
	s.eventMu.Lock()
	handlers := append([]func(){}, s.eventHandlers...)
	s.eventMu.Unlock()
	for _, h := range handlers {
		go h()
	}
}

// startEventListener maintains a persistent websocket listening for global events
func (s *Service) startEventListener() {
	s.eventMu.Lock()
	// avoid starting multiple listeners
	if s.eventClient != nil {
		s.eventMu.Unlock()
		return
	}
	s.eventMu.Unlock()

	s.mu.RLock()
	token := s.api.token
	serverURL := s.config.ServerURL
	s.mu.RUnlock()

	ctx, cancel := context.WithCancel(context.Background())
	s.eventMu.Lock()
	s.eventCancel = cancel
	s.eventMu.Unlock()

	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ws := NewWSClient(wsHostFromServerURL(serverURL), token)
		if err := ws.Connect(); err != nil {
			if !sleepOrDone(ctx, backoff) {
				return
			}
			if backoff < 5*time.Second {
				backoff *= 2
			}
			continue
		}

		s.eventMu.Lock()
		s.eventClient = ws
		s.eventMu.Unlock()

		receive := ws.ReceiveMessages()
		for {
			select {
			case <-ctx.Done():
				ws.Disconnect()
				return
			case msg := <-receive:
				if msg.Type == "friend_accepted" || msg.Type == "friend_request" || msg.Type == "friend_key_updated" {
					s.callEventHandlers()
				}
			}
		}
	}
}

func (s *Service) stopEventListener() {
	s.eventMu.Lock()
	cancel := s.eventCancel
	client := s.eventClient
	s.eventCancel = nil
	s.eventClient = nil
	s.eventMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if client != nil {
		client.Disconnect()
	}
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
				// Handle friend-related events regardless of current active friend
				if msg.Type == "friend_accepted" || msg.Type == "friend_request" || msg.Type == "friend_key_updated" {
					// Ask UI to refresh lists
					signalStatus(status, "refresh:friends")
					continue
				}

				if msg.SenderID != friend.FriendUserID {
					continue
				}

				decrypted, err := DecryptMessage(privateKey, msg.Content)
				if err != nil {
					signalStatus(status, fmt.Sprintf("Decrypt failed: %v", err))
					continue
				}

				createdAt := msg.CreatedAt
				if createdAt.IsZero() {
					createdAt = time.Now()
				}

				message := ChatMessage{
					SenderID:   msg.SenderID,
					ReceiverID: msg.ReceiverID,
					Content:    decrypted,
					IsHeart:    msg.Type == "heart",
					Incoming:   true,
					CreatedAt:  createdAt,
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
