package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	config      ClientConfig
	api         *APIClient
	wsClient    *WSClient
	storage     *LocalStorage
	currentUser User
	localKeys   LocalKeys
)

func main() {
	config = LoadClientConfig()

	var err error
	storage, err = NewLocalStorage(config.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize local storage: %v", err)
	}

	api = NewAPIClient(config.ServerURL)

	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║   🔒 poCHATo - Secure Chat Application    ║")
	fmt.Println("║  End-to-End Encrypted Messaging System     ║")
	fmt.Println("╚════════════════════════════════════════════╝")
	fmt.Println()

	// Check for existing session
	_, token, err := storage.LoadSession()
	if err == nil {
		// Session exists, try to resume
		api.SetToken(token)
		user, err := api.GetMe()
		if err == nil {
			currentUser = user
			localKeys, _ = storage.LoadKeys()
			fmt.Printf("✓ Welcome back, %s!\n\n", user.Username)
			mainMenu()
			return
		}
	}

	// No valid session, show auth menu
	authMenu()
}

func authMenu() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("1️⃣  Register")
	fmt.Println("2️⃣  Login")
	fmt.Println("3️⃣  Exit")
	fmt.Print("\n👉 Choose an option: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		handleRegister(reader)
	case "2":
		handleLogin(reader)
	case "3":
		fmt.Println("Goodbye! 👋")
		os.Exit(0)
	default:
		fmt.Println("❌ Invalid option")
		authMenu()
	}
}

func handleRegister(reader *bufio.Reader) {
	fmt.Println("\n📝 Registration")
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	response, err := api.Register(username, email, password)
	if err != nil {
		fmt.Printf("❌ Registration failed: %v\n", err)
		authMenu()
		return
	}

	// Generate key pair
	pubKey, privKey, err := GenerateKeyPair()
	if err != nil {
		fmt.Printf("❌ Failed to generate keys: %v\n", err)
		return
	}

	// Save keys locally
	storage.SaveKeys(response.UserID, pubKey, privKey)
	storage.SaveSession(response.UserID, response.Token)

	currentUser = response.User
	localKeys = LocalKeys{
		UserID:     response.UserID,
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}

	fmt.Println("✓ Registration successful!")
	fmt.Printf("✓ User ID: %s\n", response.UserID)
	fmt.Println()
	mainMenu()
}

func handleLogin(reader *bufio.Reader) {
	fmt.Println("\n🔐 Login")
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	response, err := api.Login(username, password)
	if err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		authMenu()
		return
	}

	// Load or generate keys
	keys, err := storage.LoadKeys()
	if err != nil {
		// Generate new key pair if not found
		pubKey, privKey, err := GenerateKeyPair()
		if err != nil {
			fmt.Printf("❌ Failed to generate keys: %v\n", err)
			return
		}
		storage.SaveKeys(response.UserID, pubKey, privKey)
		localKeys = LocalKeys{
			UserID:     response.UserID,
			PublicKey:  pubKey,
			PrivateKey: privKey,
		}
	} else {
		localKeys = keys
	}

	storage.SaveSession(response.UserID, response.Token)
	currentUser = response.User

	fmt.Println("✓ Login successful!")
	fmt.Printf("✓ Welcome, %s!\n", response.User.Username)
	fmt.Println()
	mainMenu()
}

func mainMenu() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("═══════════════════════════════════════════")
	fmt.Printf("👤 Logged in as: %s\n", currentUser.Username)
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println()
	fmt.Println("1️⃣  View Friends")
	fmt.Println("2️⃣  Add Friend")
	fmt.Println("3️⃣  View Friend Requests")
	fmt.Println("4️⃣  Chat with Friend")
	fmt.Println("5️⃣  View Blocked Users")
	fmt.Println("6️⃣  Logout")
	fmt.Println("7️⃣  Exit")
	fmt.Print("\n👉 Choose an option: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		viewFriends()
	case "2":
		addFriend(reader)
	case "3":
		viewFriendRequests(reader)
	case "4":
		chatWithFriend(reader)
	case "5":
		viewBlockedUsers()
	case "6":
		logout()
	case "7":
		fmt.Println("Goodbye! 👋")
		os.Exit(0)
	default:
		fmt.Println("❌ Invalid option")
		mainMenu()
	}
}

func viewFriends() {
	friends, err := api.GetFriends()
	if err != nil {
		fmt.Printf("❌ Failed to get friends: %v\n", err)
		mainMenu()
		return
	}

	if len(friends) == 0 {
		fmt.Println("\n📭 You have no friends yet!")
	} else {
		fmt.Println("\n👥 Your Friends:")
		fmt.Println("─────────────────────────────────────────")
		for _, friend := range friends {
			status := "⚪ Offline"
			fmt.Printf("• %s (%s)\n", friend.FriendUserID, status)
		}
	}

	fmt.Println()
	mainMenu()
}

func addFriend(reader *bufio.Reader) {
	fmt.Print("\n👤 Enter friend's username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	err := api.AddFriend(username)
	if err != nil {
		fmt.Printf("❌ Failed to add friend: %v\n", err)
	} else {
		fmt.Println("✓ Friend request sent!")
	}

	fmt.Println()
	mainMenu()
}

func viewFriendRequests(reader *bufio.Reader) {
	requests, err := api.GetFriendRequests()
	if err != nil {
		fmt.Printf("❌ Failed to get friend requests: %v\n", err)
		mainMenu()
		return
	}

	if len(requests) == 0 {
		fmt.Println("\n📭 No pending friend requests!")
		mainMenu()
		return
	}

	fmt.Println("\n📬 Friend Requests:")
	fmt.Println("─────────────────────────────────────────")
	for i, req := range requests {
		fmt.Printf("%d. From: %s\n", i+1, req.SenderID)
	}

	fmt.Print("\n👉 Accept request number (0 to skip): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "0" {
		mainMenu()
		return
	}

	// Parse choice and accept request
	for i, req := range requests {
		if fmt.Sprintf("%d", i+1) == choice {
			err := api.AcceptFriendRequest(req.ID)
			if err != nil {
				fmt.Printf("❌ Failed to accept request: %v\n", err)
			} else {
				fmt.Println("✓ Friend request accepted!")

				// Exchange public keys
				pubKey, _, _ := GenerateKeyPair()
				api.UpdatePublicKey(req.SenderID, pubKey)
				storage.SaveFriendPublicKey(req.SenderID, pubKey)
			}
			break
		}
	}

	fmt.Println()
	mainMenu()
}

func chatWithFriend(reader *bufio.Reader) {
	friends, err := api.GetFriends()
	if err != nil || len(friends) == 0 {
		fmt.Println("❌ No friends available")
		mainMenu()
		return
	}

	fmt.Println("\n👥 Select Friend:")
	for i, friend := range friends {
		fmt.Printf("%d. %s\n", i+1, friend.FriendUserID)
	}

	fmt.Print("\n👉 Enter friend number (0 to skip): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "0" {
		mainMenu()
		return
	}

	var selectedFriend Friend
	for i, friend := range friends {
		if fmt.Sprintf("%d", i+1) == choice {
			selectedFriend = friend
			break
		}
	}

	if selectedFriend.FriendUserID == "" {
		fmt.Println("❌ Invalid selection")
		mainMenu()
		return
	}

	// Start WebSocket connection
	wsURL := strings.TrimPrefix(config.ServerURL, "http://")
	wsClient = NewWSClient(wsURL, api.token)

	if err := wsClient.Connect(); err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		mainMenu()
		return
	}

	fmt.Printf("\n💬 Chat with %s (Type 'exit' to quit)\n", selectedFriend.FriendUserID)
	fmt.Println("─────────────────────────────────────────")

	// Load message history
	history, _ := api.GetMessageHistory(selectedFriend.FriendUserID, 10, 0)
	if len(history) > 0 {
		fmt.Println("📜 Recent Messages:")
		for _, msg := range history {
			prefix := "You:"
			if msg.SenderID != currentUser.ID {
				prefix = "Them:"
			}
			emoji := "💬"
			if msg.IsHeart {
				emoji = "❤️"
			}
			fmt.Printf("%s %s %s\n", emoji, prefix, msg.Content)
		}
		fmt.Println("─────────────────────────────────────────")
	}

	// Chat input loop
	go handleIncomingMessages(selectedFriend.FriendUserID)

	for {
		fmt.Print("You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			wsClient.Disconnect()
			mainMenu()
			return
		}

		if input == "" {
			continue
		}

		// Check if heart message
		isHeart := strings.ToLower(input) == "❤️" || strings.ToLower(input) == "heart"

		// Encrypt message
		encryptedMsg, err := EncryptMessage(selectedFriend.PublicKey, input)
		if err != nil {
			fmt.Printf("❌ Encryption failed: %v\n", err)
			continue
		}

		// Send via WebSocket
		wsClient.SendMessage(selectedFriend.FriendUserID, encryptedMsg, isHeart)
		fmt.Println("✓ Message sent")
	}
}

func handleIncomingMessages(friendID string) {
	for msg := range wsClient.ReceiveMessages() {
		if msg.SenderID == friendID {
			// Decrypt message
			decrypted, err := DecryptMessage(localKeys.PrivateKey, msg.Content)
			if err != nil {
				fmt.Printf("❌ Decryption failed: %v\n", err)
				continue
			}

			emoji := "💬"
			if msg.Type == "heart" {
				emoji = "❤️"
			}

			fmt.Printf("\n%s Them: %s\n", emoji, decrypted)
			fmt.Print("You: ")
		}
	}
}

func viewBlockedUsers() {
	blocked, err := api.GetBlockedUsers()
	if err != nil || len(blocked) == 0 {
		fmt.Println("\n✓ You haven't blocked anyone!")
	} else {
		fmt.Println("\n🚫 Blocked Users:")
		fmt.Println("─────────────────────────────────────────")
		for _, user := range blocked {
			fmt.Printf("• %v\n", user)
		}
	}

	fmt.Println()
	mainMenu()
}

func logout() {
	storage.ClearSession()
	api = NewAPIClient(config.ServerURL)
	currentUser = User{}
	fmt.Println("✓ Logged out successfully!\n")
	authMenu()
}
