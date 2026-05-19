package main

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	return createTables()
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS friend_requests (
		id TEXT PRIMARY KEY,
		sender_id TEXT NOT NULL,
		receiver_id TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(sender_id) REFERENCES users(id),
		FOREIGN KEY(receiver_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS friends (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		friend_user_id TEXT NOT NULL,
		public_key TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(friend_user_id) REFERENCES users(id),
		UNIQUE(user_id, friend_user_id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		sender_id TEXT NOT NULL,
		receiver_id TEXT NOT NULL,
		content TEXT NOT NULL,
		is_heart INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(sender_id) REFERENCES users(id),
		FOREIGN KEY(receiver_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS blocked_users (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		blocked_user_id TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(blocked_user_id) REFERENCES users(id),
		UNIQUE(user_id, blocked_user_id)
	);

	CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);
	CREATE INDEX IF NOT EXISTS idx_messages_receiver ON messages(receiver_id);
	CREATE INDEX IF NOT EXISTS idx_friends_user ON friends(user_id);
	CREATE INDEX IF NOT EXISTS idx_friend_requests_receiver ON friend_requests(receiver_id);
	`

	_, err := db.Exec(schema)
	return err
}

// User operations
func CreateUser(username, email, passwordHash string) (User, error) {
	id := uuid.New().String()
	now := time.Now()

	user := User{
		ID:        id,
		Username:  username,
		Email:     email,
		Password:  passwordHash,
		CreatedAt: now,
	}

	_, err := db.Exec(
		"INSERT INTO users (id, username, email, password, created_at) VALUES (?, ?, ?, ?, ?)",
		id, username, email, passwordHash, now,
	)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func GetUserByUsername(username string) (User, error) {
	var user User
	err := db.QueryRow(
		"SELECT id, username, email, password, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)

	return user, err
}

func GetUserByID(id string) (User, error) {
	var user User
	err := db.QueryRow(
		"SELECT id, username, email, password, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)

	return user, err
}

// Friend operations
func CreateFriend(userID, friendUserID, publicKey string) error {
	id := uuid.New().String()
	now := time.Now()

	_, err := db.Exec(
		"INSERT INTO friends (id, user_id, friend_user_id, public_key, created_at) VALUES (?, ?, ?, ?, ?)",
		id, userID, friendUserID, publicKey, now,
	)
	return err
}

func GetFriend(userID, friendUserID string) (Friend, error) {
	var friend Friend
	err := db.QueryRow(
		"SELECT id, user_id, friend_user_id, public_key, created_at FROM friends WHERE user_id = ? AND friend_user_id = ?",
		userID, friendUserID,
	).Scan(&friend.ID, &friend.UserID, &friend.FriendUserID, &friend.PublicKey, &friend.CreatedAt)

	return friend, err
}

func GetFriendsForUser(userID string) ([]Friend, error) {
	rows, err := db.Query(
		"SELECT id, user_id, friend_user_id, public_key, created_at FROM friends WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []Friend
	for rows.Next() {
		var f Friend
		if err := rows.Scan(&f.ID, &f.UserID, &f.FriendUserID, &f.PublicKey, &f.CreatedAt); err != nil {
			return nil, err
		}
		friends = append(friends, f)
	}

	return friends, rows.Err()
}

func UpdateFriendPublicKey(userID, friendUserID, publicKey string) error {
	_, err := db.Exec(
		"UPDATE friends SET public_key = ? WHERE user_id = ? AND friend_user_id = ?",
		publicKey, userID, friendUserID,
	)
	return err
}

// Friend request operations
func CreateFriendRequest(senderID, receiverID string) error {
	id := uuid.New().String()
	now := time.Now()

	_, err := db.Exec(
		"INSERT INTO friend_requests (id, sender_id, receiver_id, status, created_at, updated_at) VALUES (?, ?, ?, 'pending', ?, ?)",
		id, senderID, receiverID, now, now,
	)
	return err
}

func GetFriendRequestsPending(userID string) ([]FriendRequest, error) {
	rows, err := db.Query(
		"SELECT id, sender_id, receiver_id, status, created_at, updated_at FROM friend_requests WHERE receiver_id = ? AND status = 'pending'",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []FriendRequest
	for rows.Next() {
		var req FriendRequest
		if err := rows.Scan(&req.ID, &req.SenderID, &req.ReceiverID, &req.Status, &req.CreatedAt, &req.UpdatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, rows.Err()
}

func AcceptFriendRequest(requestID, senderID, receiverID string) error {
	now := time.Now()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Update friend request status
	_, err = tx.Exec(
		"UPDATE friend_requests SET status = 'accepted', updated_at = ? WHERE id = ?",
		now, requestID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Create bidirectional friend relationship
	id1 := uuid.New().String()
	id2 := uuid.New().String()

	_, err = tx.Exec(
		"INSERT INTO friends (id, user_id, friend_user_id, created_at) VALUES (?, ?, ?, ?)",
		id1, receiverID, senderID, now,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO friends (id, user_id, friend_user_id, created_at) VALUES (?, ?, ?, ?)",
		id2, senderID, receiverID, now,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func RejectFriendRequest(requestID string) error {
	now := time.Now()
	_, err := db.Exec(
		"UPDATE friend_requests SET status = 'rejected', updated_at = ? WHERE id = ?",
		now, requestID,
	)
	return err
}

// Message operations
func StoreMessage(senderID, receiverID, content string, isHeart bool) error {
	id := uuid.New().String()
	now := time.Now()

	_, err := db.Exec(
		"INSERT INTO messages (id, sender_id, receiver_id, content, is_heart, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, senderID, receiverID, content, isHeart, now,
	)
	return err
}

func GetMessageHistory(userID, friendID string, limit, offset int) ([]Message, error) {
	rows, err := db.Query(`
		SELECT id, sender_id, receiver_id, content, is_heart, created_at
		FROM messages
		WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, friendID, friendID, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		var isHeartInt int
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &isHeartInt, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.IsHeart = isHeartInt == 1
		messages = append(messages, m)
	}

	return messages, rows.Err()
}

// Block operations
func BlockUser(userID, blockedUserID string) error {
	id := uuid.New().String()
	now := time.Now()

	_, err := db.Exec(
		"INSERT INTO blocked_users (id, user_id, blocked_user_id, created_at) VALUES (?, ?, ?, ?)",
		id, userID, blockedUserID, now,
	)
	return err
}

func UnblockUser(userID, blockedUserID string) error {
	_, err := db.Exec(
		"DELETE FROM blocked_users WHERE user_id = ? AND blocked_user_id = ?",
		userID, blockedUserID,
	)
	return err
}

func IsUserBlocked(userID, blockedUserID string) (bool, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM blocked_users WHERE user_id = ? AND blocked_user_id = ?",
		userID, blockedUserID,
	).Scan(&count)

	return count > 0, err
}

func GetBlockedUsers(userID string) ([]BlockedUser, error) {
	rows, err := db.Query(
		"SELECT id, user_id, blocked_user_id, created_at FROM blocked_users WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocked []BlockedUser
	for rows.Next() {
		var b BlockedUser
		if err := rows.Scan(&b.ID, &b.UserID, &b.BlockedUserID, &b.CreatedAt); err != nil {
			return nil, err
		}
		blocked = append(blocked, b)
	}

	return blocked, rows.Err()
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
