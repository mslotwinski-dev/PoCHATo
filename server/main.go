package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	config := LoadConfig()

	// Initialize database
	if err := InitDB(config.DatabasePath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()

	// Initialize WebSocket client manager
	InitClientManager()

	// Register HTTP routes
	registerRoutes()

	fmt.Printf("🚀 poCHATo Server running on port %s\n", config.ServerPort)
	fmt.Println("✨ Secure, Real-Time, End-to-End Encrypted Chat")
	fmt.Println("📝 Database: " + config.DatabasePath)
	fmt.Println()

	err := http.ListenAndServe(config.ServerPort, nil)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func registerRoutes() {
	// Authentication routes
	http.HandleFunc("/api/auth/register", RegisterHandler)
	http.HandleFunc("/api/auth/login", LoginHandler)
	http.HandleFunc("/api/auth/logout", LogoutHandler)
	http.HandleFunc("/api/auth/me", GetMeHandler)

	// Friend routes
	http.HandleFunc("/api/friends/add", AddFriendHandler)
	http.HandleFunc("/api/friends/requests", GetFriendRequestsHandler)
	http.HandleFunc("/api/friends/accept", AcceptFriendHandler)
	http.HandleFunc("/api/friends/list", GetFriendsHandler)
	http.HandleFunc("/api/friends/key", UpdatePublicKeyHandler)
	http.HandleFunc("/api/friends/history", GetMessageHistoryHandler)

	// Block routes
	http.HandleFunc("/api/block/user", BlockUserHandler)
	http.HandleFunc("/api/block/unblock", UnblockUserHandler)
	http.HandleFunc("/api/block/list", GetBlockedUsersHandler)

	// WebSocket route
	http.HandleFunc("/ws", WebSocketHandler)

	// Health check
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"healthy"}`)
	})
}
