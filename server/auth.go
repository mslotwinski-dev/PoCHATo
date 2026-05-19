package main

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// SimpleTokenMap stores tokens in memory (in production, use JWT or Redis)
var activeTokens = make(map[string]string) // token -> userID

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func GenerateToken(userID string) string {
	token := uuid.New().String() + "-" + uuid.New().String()
	activeTokens[token] = userID
	return token
}

func ValidateToken(token string) (string, error) {
	userID, exists := activeTokens[token]
	if !exists {
		return "", errors.New("invalid token")
	}
	return userID, nil
}

func RevokeToken(token string) {
	delete(activeTokens, token)
}

// RegisterUser creates a new user account
func RegisterUser(username, email, password string) (User, error) {
	// Check if user already exists
	_, err := GetUserByUsername(username)
	if err == nil {
		return User{}, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}

	// Create user in database
	user, err := CreateUser(username, email, hashedPassword)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// LoginUser authenticates a user
func LoginUser(username, password string) (User, string, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		return User{}, "", errors.New("invalid credentials")
	}

	if err := VerifyPassword(user.Password, password); err != nil {
		return User{}, "", errors.New("invalid credentials")
	}

	// Generate token
	token := GenerateToken(user.ID)

	// Clear password from response
	user.Password = ""

	return user, token, nil
}
