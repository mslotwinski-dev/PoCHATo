<div align="center">
  <img src="https://minecraft.wiki/images/Potato_JE3_BE2.png" alt="poCHATo Logo" width="150"/>

  # poCHATo
  
  **A secure, real-time, end-to-end encrypted client-server chat application written in Go.**

  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
  [![Gin Framework](https://img.shields.io/badge/Gin-Web_Framework-0088CC?style=flat)](https://gin-gonic.com/)
  [![WebSockets](https://img.shields.io/badge/WebSockets-Real_Time-blue?style=flat)](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

---

## 📖 Overview

**poCHATo** is a comprehensive computer networks project designed to provide secure, real-time messaging between users. Built entirely in Go, it features a robust client-server architecture utilizing a RESTful API for standard operations and WebSockets (TCP/IP) for low-latency, bi-directional communication. 

Security and privacy are at the core of poCHATo. Direct communication is restricted to mutually accepted friends, and every single message is secured using **End-to-End Encryption (E2EE)**. 

## ✨ Key Features

* **🔒 End-to-End Encryption (E2EE):** Total privacy. Public and private key pairs (RSA 2048-bit) are generated when users become friends. Messages are fully encrypted on the sender's device and decrypted only on the recipient's device. The server never reads the plaintext messages.
* **🤝 Friend System:** Users must establish a connection before chatting seamlessly. You can easily fetch your friend list and manage your connections. Automatic public key exchange on acceptance.
* **⚡ Real-Time WebSockets:** Powered by TCP/IP and WebSockets for instant message delivery without continuous HTTP polling.
* **💬 Rich Chat Interactions:** 
    * Text messages
    * Heart messages (❤️) for quick reactions
    * Typing indicators
    * Message history retrieval
* **🚫 User Blocking:** Block malicious or unwanted users. Prevent blocked users from sending messages.
* **🔐 Secure Authentication:** Robust user registration and login system with bcrypt-hashed passwords. Token-based session management.

---

## 🏗️ Architecture & Tech Stack

poCHATo consists of a backend server and a client application, strictly separating concerns:

### Backend (Server)
* **Language:** Go 1.24.2
* **API Framework:** Net/http for fast RESTful routing (Login, Registration, Friend Management)
* **Real-time Protocol:** Gorilla WebSockets over TCP/IP for active chat channels
* **Database:** SQLite3 for storing users, hashed credentials, and encrypted messages
* **Encryption:** crypto/rsa for E2EE, bcrypt for password hashing

### Frontend (Client)
* **Language:** Go 1.24.2
* **API Client:** HTTP client for REST API communication
* **WebSocket Client:** Gorilla WebSockets for real-time messaging
* **Cryptography:** crypto/rsa for E2EE, local JSON-based storage for keys
* **Interface:** Fyne desktop GUI with native windowing on Windows, Linux, and macOS

---

## 🔐 How E2EE Works in poCHATo

To ensure the server cannot read messages, poCHATo implements a strict cryptographic flow:

1. **Friend Request Accepted:** When User A and User B become friends, their clients generate standard RSA Public/Private key pairs (2048-bit).
2. **Key Exchange:** Public keys are exchanged through the server and stored locally on the clients. Private keys *never* leave the user's local device.
3. **Sending a Message:** When User A messages User B, User A's client encrypts the message payload using User B's Public Key with OAEP padding.
4. **Storage & Delivery:** The server receives and stores the ciphertext, delivering it via WebSockets to User B.
5. **Decryption:** User B's client receives the ciphertext and decrypts it using User B's Private Key, revealing the original message.

---

## 🚀 Getting Started

### Prerequisites
* [Go](https://go.dev/dl/) 1.24.2 or higher
* SQLite3 (included via go-sqlite3)
* Git for cloning

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/poCHATo.git
   cd poCHATo
   ```

2. **Download dependencies:**
   ```bash
   go mod download
   ```

3. **Create .env file (optional):**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

### Running the Server

```bash
go run ./server
```

Expected output:
```
🚀 poCHATo Server running on port :8080
✨ Secure, Real-Time, End-to-End Encrypted Chat
📝 Database: ./pochato.db
```

### Running the Desktop Client

In a different terminal:
```bash
go run .
```

You'll see the native Fyne desktop app with authentication, friend lists, and chat.

---

## 📡 API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user  
- `POST /api/auth/logout` - Logout user
- `GET /api/auth/me` - Get current user info

### Friends
- `POST /api/friends/add` - Send friend request
- `GET /api/friends/requests` - Get pending friend requests
- `POST /api/friends/accept` - Accept friend request
- `GET /api/friends/list` - Get friends list
- `POST /api/friends/key` - Update friend's public key
- `GET /api/friends/history` - Get message history

### Blocking
- `POST /api/block/user` - Block a user
- `POST /api/block/unblock` - Unblock a user
- `GET /api/block/list` - Get blocked users list

### WebSocket
- `WS /ws?token=<token>` - WebSocket connection for real-time messaging

### Health
- `GET /api/health` - Server health check

---

## 🗂️ Project Structure

```
poCHATo/
├── main.go                  # Desktop GUI entry point
├── desktop_ui.go            # Fyne UI wiring
├── server/                  # Backend server
│   ├── main.go              # Entry point & route registration
│   ├── config.go            # Configuration management
│   ├── models.go            # Data structures
│   ├── database.go          # SQLite3 operations
│   ├── auth.go              # Authentication & password hashing
│   ├── encryption.go        # RSA encryption utilities
│   ├── websocket.go         # WebSocket client management
│   ├── handlers.go          # HTTP request handlers
│   └── errors.go            # Error definitions
│
├── app/                     # Reusable client package (auth, storage, transport)
│   ├── main.go              # Legacy CLI helpers kept for reference
│   ├── config.go            # Configuration management
│   ├── models.go            # Data structures
│   ├── api.go               # REST API client
│   ├── websocket.go         # WebSocket client
│   ├── encryption.go        # RSA encryption/decryption
│   ├── service.go           # Client service layer for the GUI
│   └── storage.go           # Local JSON-based storage
│
├── go.mod                   # Go module definition
go run ./server
├── LICENSE                  # MIT License
└── README.md                # This file
```

---

## 🔄 Data Flow
go run .
```
poCHATo desktop window
```
Login, register, friend management, and chat all run in the native GUI.
### Friend Request Flow
The client stores session and key material under the configured `DATA_DIR`.

### Encryption
- **RSA 2048-bit** for asymmetric encryption
- **OAEP padding** for secure encryption
- **Base64 encoding** for safe transmission
- **Zero-knowledge server:** Never stores plaintext messages

### Password Security
- **Bcrypt hashing** with configurable cost factor
- **Never stored in plaintext**
- **Constant-time comparison** for verification

### Communication
- **HTTPS/WSS recommended** in production
- **Token-based authentication** instead of credentials
- **CORS support** for development

### Local Storage
- **Keys stored with 0600 permissions** (read/write owner only)
- **Session data stored locally** (never transmitted)
- **Private keys never sent to server**

---

## 💻 Usage Example

### Server
```bash
$ cd server
$ go run .
🚀 poCHATo Server running on port :8080
✨ Secure, Real-Time, End-to-End Encrypted Chat
📝 Database: ./pochato.db
```

### Client Session
```bash
$ cd app
$ go run .

╔════════════════════════════════════════════╗
║   🔒 poCHATo - Secure Chat Application    ║
║  End-to-End Encrypted Messaging System     ║
╚════════════════════════════════════════════╝

1️⃣  Register
2️⃣  Login
3️⃣  Exit

👉 Choose an option: 1

📝 Registration
Username: alice
Email: alice@example.com
Password: ••••••••
✓ Registration successful!

👤 Logged in as: alice
═══════════════════════════════════════════

1️⃣  View Friends
2️⃣  Add Friend
3️⃣  View Friend Requests
4️⃣  Chat with Friend
5️⃣  View Blocked Users
6️⃣  Logout
7️⃣  Exit
```

---

## 🧪 Testing with curl

### Register
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'
```

### Get Current User
```bash
curl -H "Authorization: <token>" \
  http://localhost:8080/api/auth/me
```

### Health Check
```bash
curl http://localhost:8080/api/health
```

---

## 📊 Database Schema

### Users Table
- `id` - Primary key (UUID)
- `username` - Unique username
- `email` - User email
- `password` - Bcrypt hash
- `created_at` - Registration timestamp

### Friends Table
- `id` - Primary key (UUID)
- `user_id` - Foreign key to users
- `friend_user_id` - Foreign key to users
- `public_key` - Friend's public key (base64)
- `created_at` - Friendship creation timestamp

### Messages Table
- `id` - Primary key (UUID)
- `sender_id` - Foreign key to users
- `receiver_id` - Foreign key to users
- `content` - Encrypted message (base64)
- `is_heart` - Heart reaction flag
- `created_at` - Message timestamp

### Blocked Users Table
- `id` - Primary key (UUID)
- `user_id` - Foreign key to users
- `blocked_user_id` - Foreign key to users
- `created_at` - Block timestamp

---

## 🤝 Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## ⚠️ Disclaimer

This is an educational project. While it implements E2EE, please conduct a security audit before using in production environments. Always follow security best practices:
- Use HTTPS/WSS in production
- Implement proper key management
- Regular security updates
- Compliance with data protection regulations (GDPR, CCPA, etc.)

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 👨‍💻 Authors

**Mateusz Słotwiński**
**Łukasz Przeszłowski**

For questions or support, please open an issue on GitHub.

---

## 🔗 Sources

- **Repository:** https://github.com/mslotwinski-dev/poCHATo
- **Go Documentation:** https://golang.org/
- **Gorilla WebSocket:** https://github.com/gorilla/websocket
- **SQLite:** https://sqlite.org/
- **Adam Woodbeck - Network Programming with Go:** https://github.com/irezaul/go-life/tree/main/books

---

**Made with ❤️ for secure communication**
