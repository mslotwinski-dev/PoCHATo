package main

import "time"

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// FriendRequest represents a friend request between two users
type FriendRequest struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Status     string    `json:"status"` // pending, accepted, rejected
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Friend represents a friendship between two users
type Friend struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	FriendUserID string    `json:"friend_user_id"`
	PublicKey    string    `json:"public_key"`
	CreatedAt    time.Time `json:"created_at"`
}

// Message represents a chat message
type Message struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Content    string    `json:"content"` // Encrypted content
	IsHeart    bool      `json:"is_heart"`
	CreatedAt  time.Time `json:"created_at"`
}

// BlockedUser represents a user that has been blocked
type BlockedUser struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	BlockedUserID string    `json:"blocked_user_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// WSMessage represents a message sent over WebSocket
type WSMessage struct {
	Type       string `json:"type"` // message, heart, typing, etc.
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
}

// AuthRequest represents a login/register request
type AuthRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents a login/register response
type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	User   User   `json:"user"`
}

// KeyExchangeRequest represents a request to exchange public keys
type KeyExchangeRequest struct {
	FriendID  string `json:"friend_id"`
	PublicKey string `json:"public_key"`
}

// MessageHistoryRequest represents a request for message history
type MessageHistoryRequest struct {
	FriendID string `json:"friend_id"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}
