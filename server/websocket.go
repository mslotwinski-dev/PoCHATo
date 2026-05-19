package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ClientManager manages WebSocket connections
type ClientManager struct {
	clients    map[string]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Client represents a connected WebSocket client
type Client struct {
	ID      string
	conn    *websocket.Conn
	send    chan interface{}
	manager *ClientManager
}

var clientManager *ClientManager

func InitClientManager() {
	clientManager = &ClientManager{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	go clientManager.Run()
}

// Run manages client connections and message broadcasting
func (cm *ClientManager) Run() {
	for {
		select {
		case client := <-cm.register:
			cm.mu.Lock()
			cm.clients[client.ID] = client
			cm.mu.Unlock()
			log.Printf("Client registered: %s (total: %d)", client.ID, len(cm.clients))

		case client := <-cm.unregister:
			cm.mu.Lock()
			if _, exists := cm.clients[client.ID]; exists {
				delete(cm.clients, client.ID)
				close(client.send)
			}
			cm.mu.Unlock()
			log.Printf("Client unregistered: %s (total: %d)", client.ID, len(cm.clients))

		case message := <-cm.broadcast:
			cm.mu.RLock()
			for _, client := range cm.clients {
				select {
				case client.send <- message:
				default:
					// Client send channel full, skip
				}
			}
			cm.mu.RUnlock()
		}
	}
}

// SendToClient sends a message to a specific client
func (cm *ClientManager) SendToClient(userID string, message interface{}) {
	cm.mu.RLock()
	client, exists := cm.clients[userID]
	cm.mu.RUnlock()

	if exists {
		select {
		case client.send <- message:
		default:
			log.Printf("Failed to send message to client %s (channel full)", userID)
		}
	}
}

// GetClient retrieves a client by ID
func (cm *ClientManager) GetClient(userID string) *Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clients[userID]
}

// IsClientConnected checks if a client is connected
func (cm *ClientManager) IsClientConnected(userID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, exists := cm.clients[userID]
	return exists
}

// NewClient creates a new client
func NewClient(id string, conn *websocket.Conn, manager *ClientManager) *Client {
	return &Client{
		ID:      id,
		conn:    conn,
		send:    make(chan interface{}, 256),
		manager: manager,
	}
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Time{})

	for {
		var msg WSMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		msg.SenderID = c.ID
		c.handleMessage(&msg)
	}
}

// WritePump sends messages to the WebSocket connection
func (c *Client) WritePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Time{})
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		default:
			// Continue if send channel is empty
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(msg *WSMessage) {
	switch msg.Type {
	case "message", "heart":
		// Store message in database
		isHeart := msg.Type == "heart"
		err := StoreMessage(c.ID, msg.ReceiverID, msg.Content, isHeart)
		if err != nil {
			log.Printf("Failed to store message: %v", err)
			return
		}

		// Check if receiver is blocked by sender
		blocked, err := IsUserBlocked(msg.ReceiverID, c.ID)
		if err != nil {
			log.Printf("Error checking block status: %v", err)
			return
		}

		if blocked {
			log.Printf("Message blocked: sender %s is blocked by receiver %s", c.ID, msg.ReceiverID)
			return
		}

		// Send to receiver if connected
		c.manager.SendToClient(msg.ReceiverID, msg)

	case "typing":
		// Send typing indicator to receiver
		if msg.ReceiverID != "" {
			c.manager.SendToClient(msg.ReceiverID, msg)
		}

	case "acknowledgment":
		// Handle message acknowledgment (optional)
		log.Printf("Message acknowledged by %s", c.ID)

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true // Allow all origins for development
// 	},
// }

// func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
// 	userID := r.URL.Query().Get("user_id")
// 	if userID == "" {
// 		http.Error(w, "Missing user_id", http.StatusBadRequest)
// 		return
// 	}

// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Printf("WebSocket upgrade error: %v", err)
// 		return
// 	}

// 	client := NewClient(userID, conn, clientManager)
// 	clientManager.register <- client

// 	go client.WritePump()
// 	go client.ReadPump()
// }
