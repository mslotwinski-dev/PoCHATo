package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// SetToken sets the authentication token
func (ac *APIClient) SetToken(token string) {
	ac.token = token
}

// Register registers a new user
func (ac *APIClient) Register(username, email, password string) (AuthResponse, error) {
	var response AuthResponse

	payload := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ac.baseURL+"/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return response, fmt.Errorf("registration failed: %s", string(bodyBytes))
	}

	json.NewDecoder(resp.Body).Decode(&response)
	ac.token = response.Token

	return response, nil
}

// Login authenticates a user
func (ac *APIClient) Login(username, password string) (AuthResponse, error) {
	var response AuthResponse

	payload := map[string]string{
		"username": username,
		"password": password,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ac.baseURL+"/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return response, fmt.Errorf("login failed: %s", string(bodyBytes))
	}

	json.NewDecoder(resp.Body).Decode(&response)
	ac.token = response.Token

	return response, nil
}

// GetMe retrieves current user information
func (ac *APIClient) GetMe() (User, error) {
	var user User

	req, _ := http.NewRequest("GET", ac.baseURL+"/api/auth/me", nil)
	req.Header.Set("Authorization", ac.token)

	resp, err := ac.client.Do(req)
	if err != nil {
		return user, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return user, fmt.Errorf("failed to get user info")
	}

	json.NewDecoder(resp.Body).Decode(&user)
	return user, nil
}

// AddFriend sends a friend request
func (ac *APIClient) AddFriend(username string) error {
	payload := map[string]string{"username": username}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", ac.baseURL+"/api/friends/add", bytes.NewBuffer(body))
	req.Header.Set("Authorization", ac.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add friend: %s", string(bodyBytes))
	}

	return nil
}

// GetFriendRequests retrieves pending friend requests
func (ac *APIClient) GetFriendRequests() ([]FriendRequest, error) {
	var requests []FriendRequest

	req, _ := http.NewRequest("GET", ac.baseURL+"/api/friends/requests", nil)
	req.Header.Set("Authorization", ac.token)

	resp, err := ac.client.Do(req)
	if err != nil {
		return requests, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return requests, fmt.Errorf("failed to get friend requests")
	}

	json.NewDecoder(resp.Body).Decode(&requests)
	return requests, nil
}

// AcceptFriendRequest accepts a friend request
func (ac *APIClient) AcceptFriendRequest(requestID string) error {
	payload := map[string]string{"request_id": requestID}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", ac.baseURL+"/api/friends/accept", bytes.NewBuffer(body))
	req.Header.Set("Authorization", ac.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to accept friend request")
	}

	return nil
}

// GetFriends retrieves user's friend list
func (ac *APIClient) GetFriends() ([]Friend, error) {
	var friends []Friend

	req, _ := http.NewRequest("GET", ac.baseURL+"/api/friends/list", nil)
	req.Header.Set("Authorization", ac.token)

	resp, err := ac.client.Do(req)
	if err != nil {
		return friends, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return friends, fmt.Errorf("failed to get friends list")
	}

	json.NewDecoder(resp.Body).Decode(&friends)
	return friends, nil
}

// UpdatePublicKey updates a friend's public key
func (ac *APIClient) UpdatePublicKey(friendID, publicKey string) error {
	payload := map[string]string{
		"friend_id":  friendID,
		"public_key": publicKey,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", ac.baseURL+"/api/friends/key", bytes.NewBuffer(body))
	req.Header.Set("Authorization", ac.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update public key")
	}

	return nil
}

// GetMessageHistory retrieves message history with a friend
func (ac *APIClient) GetMessageHistory(friendID string, limit, offset int) ([]Message, error) {
	var messages []Message

	url := fmt.Sprintf("%s/api/friends/history?friend_id=%s&limit=%d&offset=%d",
		ac.baseURL, friendID, limit, offset)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", ac.token)

	resp, err := ac.client.Do(req)
	if err != nil {
		return messages, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return messages, fmt.Errorf("failed to get message history")
	}

	json.NewDecoder(resp.Body).Decode(&messages)
	return messages, nil
}

// BlockUser blocks a user
func (ac *APIClient) BlockUser(blockedUserID string) error {
	payload := map[string]string{"blocked_user_id": blockedUserID}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", ac.baseURL+"/api/block/user", bytes.NewBuffer(body))
	req.Header.Set("Authorization", ac.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to block user")
	}

	return nil
}

// UnblockUser unblocks a user
func (ac *APIClient) UnblockUser(blockedUserID string) error {
	payload := map[string]string{"blocked_user_id": blockedUserID}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", ac.baseURL+"/api/block/unblock", bytes.NewBuffer(body))
	req.Header.Set("Authorization", ac.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to unblock user")
	}

	return nil
}

// GetBlockedUsers retrieves list of blocked users
func (ac *APIClient) GetBlockedUsers() ([]interface{}, error) {
	var blocked []interface{}

	req, _ := http.NewRequest("GET", ac.baseURL+"/api/block/list", nil)
	req.Header.Set("Authorization", ac.token)

	resp, err := ac.client.Do(req)
	if err != nil {
		return blocked, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return blocked, fmt.Errorf("failed to get blocked users")
	}

	json.NewDecoder(resp.Body).Decode(&blocked)
	return blocked, nil
}
