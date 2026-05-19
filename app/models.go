package main

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Friend represents a friend with public key
type Friend struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	FriendUserID string    `json:"friend_user_id"`
	PublicKey    string    `json:"public_key"`
	CreatedAt    time.Time `json:"created_at"`
}

// FriendRequest represents a friend request
type FriendRequest struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Message represents a chat message
type Message struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Content    string    `json:"content"`
	IsHeart    bool      `json:"is_heart"`
	CreatedAt  time.Time `json:"created_at"`
}

// LocalKeys stores user's RSA key pair
type LocalKeys struct {
	UserID     string
	PublicKey  string
	PrivateKey string
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type       string `json:"type"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
}

// AuthResponse from server
type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	User   User   `json:"user"`
}
