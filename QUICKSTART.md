# poCHATo - Quick Start Guide

## 🚀 Quick Setup (5 minutes)

### Prerequisites
- Go 1.24.2 or higher
- Terminal/Command Prompt

### Step 1: Clone & Setup
```bash
git clone https://github.com/yourusername/poCHATo.git
cd poCHATo
go mod download
```

### Step 2: Start the Server
```bash
go run ./server
```

You should see:
```
🚀 poCHATo Server running on port :8080
✨ Secure, Real-Time, End-to-End Encrypted Chat
📝 Database: ./pochato.db
```

### Step 3: Start the Client (New Terminal)
```bash
go run .
```

### Step 4: Create Your Account
The desktop app opens a native window with login and registration cards.

### Step 5: Add a Friend
Start another client instance with a different username:
```bash
cd app
go run .
```

Register as:
```
Username: bob
Email: bob@example.com
Password: password123
```

Back to Alice's terminal:
```
2️⃣  Add Friend
👉 Enter friend's username: bob
✓ Friend request sent!
```

### Step 6: Accept Friend Request
In Bob's terminal:
```
3️⃣  View Friend Requests
1️⃣  Accept request
✓ Friend request accepted!
```

### Step 7: Start Chatting!
Alice's terminal:
```
4️⃣  Chat with Friend
1️⃣  bob
💬 Chat with bob
You: Hello Bob!
✓ Message sent
```

Bob's terminal will receive the encrypted message and decrypt it!

---

## 🔧 Build & Run

### Build Binaries
```bash
# Build server
go build -o pochato-server ./server

# Build desktop client
go build -o pochato-desktop .
```

### Run from Binaries
```bash
# Terminal 1 - Server
./pochato-server

# Terminal 2 - Client
./pochato-desktop
```

---

## 📊 Test with curl

### Register User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "token": "abc123...",
  "user_id": "uuid...",
  "user": {
    "id": "uuid...",
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2024-05-19T..."
  }
}
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### Get Current User
```bash
curl -H "Authorization: <token-from-login>" \
  http://localhost:8080/api/auth/me
```

### Health Check
```bash
curl http://localhost:8080/api/health
```

---

## 📝 Configuration

Edit `.env` file to customize:
```bash
cp .env.example .env
```

Available options:
```
SERVER_URL=http://localhost:8080       # Client connects to this URL
DATABASE_PATH=./pochato.db             # Server database location
SERVER_PORT=:8080                      # Server listen port
JWT_SECRET=your-secret-key             # JWT secret (change in production!)
DATA_DIR=./.pochato                    # Client data directory
```

---

## 🐛 Troubleshooting

### Port Already in Use
If port 8080 is busy, change `SERVER_PORT` in `.env`:
```bash
SERVER_PORT=:8090
```

### Database Error
Delete existing database and restart:
```bash
rm pochato.db
cd server
go run .
```

### Connection Failed
Make sure server is running on correct port:
```bash
# Check server is listening
lsof -i :8080  # macOS/Linux
netstat -an | find ":8080"  # Windows
```

---

## 📚 More Information

- **Full README:** See [README.md](README.md)
- **API Documentation:** See API endpoints section in README
- **Security:** Read E2EE explanation in README

---

## 💡 Next Steps

1. **Customize the UI:** Modify `app/main.go` for different interface
2. **Deploy:** Use Docker or deploy to cloud
3. **Extend:** Add group chats, file sharing, etc.
4. **Audit:** Conduct security review before production

---

**Enjoy secure messaging! 🔐**
