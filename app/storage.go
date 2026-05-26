package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage handles local data persistence
type LocalStorage struct {
	dataDir string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(dataDir string) (*LocalStorage, error) {
	ls := &LocalStorage{
		dataDir: dataDir,
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}

	return ls, nil
}

func (ls *LocalStorage) keyFilePath(userID string) string {
	return filepath.Join(ls.dataDir, "keys", sanitizeFilePart(userID)+".json")
}

func (ls *LocalStorage) friendKeysFilePath(userID string) string {
	return filepath.Join(ls.dataDir, "friend_keys", sanitizeFilePart(userID)+".json")
}

func sanitizeFilePart(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, string(os.PathSeparator), "_")
	return value
}

// SaveKeys saves user's key pair
func (ls *LocalStorage) SaveKeys(userID string, publicKey, privateKey string) error {
	keysData := map[string]string{
		"user_id":     userID,
		"public_key":  publicKey,
		"private_key": privateKey,
	}

	data, err := json.Marshal(keysData)
	if err != nil {
		return err
	}

	keyFilePath := ls.keyFilePath(userID)
	if err := os.MkdirAll(filepath.Dir(keyFilePath), 0700); err != nil {
		return err
	}
	return os.WriteFile(keyFilePath, data, 0600)
}

// LoadKeys loads user's key pair
func (ls *LocalStorage) LoadKeys(userID string) (LocalKeys, error) {
	keyFilePath := ls.keyFilePath(userID)

	data, err := os.ReadFile(keyFilePath)
	if err != nil {
		legacyPath := filepath.Join(ls.dataDir, "keys.json")
		data, err = os.ReadFile(legacyPath)
		if err != nil {
			return LocalKeys{}, err
		}
	}

	var keysData map[string]string
	if err := json.Unmarshal(data, &keysData); err != nil {
		return LocalKeys{}, err
	}
	if userID != "" && keysData["user_id"] != "" && keysData["user_id"] != userID {
		return LocalKeys{}, fmt.Errorf("stored keys belong to another account")
	}

	return LocalKeys{
		UserID:     keysData["user_id"],
		PublicKey:  keysData["public_key"],
		PrivateKey: keysData["private_key"],
	}, nil
}

// SaveSession saves session data (user info, token)
func (ls *LocalStorage) SaveSession(userID, token string) error {
	sessionData := map[string]string{
		"user_id": userID,
		"token":   token,
	}

	data, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	sessionPath := filepath.Join(ls.dataDir, "session.json")
	return os.WriteFile(sessionPath, data, 0600)
}

// LoadSession loads session data
func (ls *LocalStorage) LoadSession() (userID, token string, err error) {
	sessionPath := filepath.Join(ls.dataDir, "session.json")

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return "", "", err
	}

	var sessionData map[string]string
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return "", "", err
	}

	return sessionData["user_id"], sessionData["token"], nil
}

// SavePreferences saves local desktop preferences.
func (ls *LocalStorage) SavePreferences(p Preferences) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	prefsPath := filepath.Join(ls.dataDir, "preferences.json")
	return os.WriteFile(prefsPath, data, 0600)
}

// LoadPreferences loads local desktop preferences.
func (ls *LocalStorage) LoadPreferences() (Preferences, error) {
	prefsPath := filepath.Join(ls.dataDir, "preferences.json")

	data, err := os.ReadFile(prefsPath)
	if err != nil {
		return Preferences{}, err
	}

	var prefs Preferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return Preferences{}, err
	}

	return prefs, nil
}

// ClearSession clears session data
func (ls *LocalStorage) ClearSession() error {
	sessionPath := filepath.Join(ls.dataDir, "session.json")
	return os.Remove(sessionPath)
}

// SaveFriendPublicKey saves a friend's public key
func (ls *LocalStorage) SaveFriendPublicKey(userID, friendID, publicKey string) error {
	friendKeysPath := ls.friendKeysFilePath(userID)
	if err := os.MkdirAll(filepath.Dir(friendKeysPath), 0700); err != nil {
		return err
	}

	var friendKeys map[string]string

	// Load existing friend keys
	if data, err := os.ReadFile(friendKeysPath); err == nil {
		json.Unmarshal(data, &friendKeys)
	}

	if friendKeys == nil {
		friendKeys = make(map[string]string)
	}

	friendKeys[friendID] = publicKey

	data, err := json.Marshal(friendKeys)
	if err != nil {
		return err
	}

	return os.WriteFile(friendKeysPath, data, 0600)
}

// GetFriendPublicKey retrieves a friend's public key
func (ls *LocalStorage) GetFriendPublicKey(userID, friendID string) (string, error) {
	friendKeysPath := ls.friendKeysFilePath(userID)

	data, err := os.ReadFile(friendKeysPath)
	if err != nil {
		return "", fmt.Errorf("friend keys file not found")
	}

	var friendKeys map[string]string
	if err := json.Unmarshal(data, &friendKeys); err != nil {
		return "", err
	}

	publicKey, exists := friendKeys[friendID]
	if !exists {
		return "", fmt.Errorf("public key for friend %s not found", friendID)
	}

	return publicKey, nil
}

// LoadAllFriendKeys loads all friend public keys
func (ls *LocalStorage) LoadAllFriendKeys(userID string) (map[string]string, error) {
	friendKeysPath := ls.friendKeysFilePath(userID)

	data, err := os.ReadFile(friendKeysPath)
	if err != nil {
		return make(map[string]string), nil
	}

	var friendKeys map[string]string
	if err := json.Unmarshal(data, &friendKeys); err != nil {
		return nil, err
	}

	return friendKeys, nil
}
