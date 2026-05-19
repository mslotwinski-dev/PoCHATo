package main

import (
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

// WSClient handles WebSocket communication
type WSClient struct {
	url         string
	token       string
	conn        *websocket.Conn
	send        chan interface{}
	receive     chan WSMessage
	isConnected bool
	mu          sync.RWMutex
	stopChan    chan struct{}
	reconnectCh chan struct{}
}

// NewWSClient creates a new WebSocket client
func NewWSClient(wsURL, token string) *WSClient {
	return &WSClient{
		url:         wsURL,
		token:       token,
		send:        make(chan interface{}, 256),
		receive:     make(chan WSMessage, 256),
		stopChan:    make(chan struct{}),
		reconnectCh: make(chan struct{}),
	}
}

// Connect establishes WebSocket connection
func (wc *WSClient) Connect() error {
	u := url.URL{Scheme: "ws", Host: wc.url, Path: "/ws"}
	q := u.Query()
	q.Set("token", wc.token)
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("websocket dial error: %v", err)
	}

	wc.mu.Lock()
	wc.conn = conn
	wc.isConnected = true
	wc.mu.Unlock()

	go wc.readPump()
	go wc.writePump()

	return nil
}

// SendMessage sends a message over WebSocket
func (wc *WSClient) SendMessage(receiverID, content string, isHeart bool) {
	msg := WSMessage{
		Type:       "message",
		ReceiverID: receiverID,
		Content:    content,
	}

	if isHeart {
		msg.Type = "heart"
	}

	select {
	case wc.send <- msg:
	default:
		log.Println("Send channel full, message dropped")
	}
}

// SendTypingIndicator sends a typing indicator
func (wc *WSClient) SendTypingIndicator(receiverID string) {
	msg := WSMessage{
		Type:       "typing",
		ReceiverID: receiverID,
	}

	select {
	case wc.send <- msg:
	default:
	}
}

// IsConnected checks if WebSocket is connected
func (wc *WSClient) IsConnected() bool {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	return wc.isConnected
}

// Disconnect closes the WebSocket connection
func (wc *WSClient) Disconnect() {
	wc.mu.Lock()
	if wc.conn != nil {
		wc.isConnected = false
		wc.conn.Close()
	}
	wc.mu.Unlock()

	select {
	case wc.stopChan <- struct{}{}:
	default:
	}
}

// readPump reads messages from the WebSocket
func (wc *WSClient) readPump() {
	defer func() {
		wc.mu.Lock()
		wc.isConnected = false
		wc.conn.Close()
		wc.mu.Unlock()
	}()

	for {
		select {
		case <-wc.stopChan:
			return
		default:
		}

		var msg WSMessage
		wc.mu.RLock()
		conn := wc.conn
		wc.mu.RUnlock()

		if conn == nil {
			return
		}

		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		select {
		case wc.receive <- msg:
		case <-wc.stopChan:
			return
		default:
			log.Println("Receive channel full, message dropped")
		}
	}
}

// writePump sends messages to the WebSocket
func (wc *WSClient) writePump() {
	defer func() {
		wc.mu.Lock()
		if wc.conn != nil {
			wc.conn.Close()
		}
		wc.mu.Unlock()
	}()

	for {
		select {
		case <-wc.stopChan:
			return
		case msg := <-wc.send:
			wc.mu.Lock()
			conn := wc.conn
			wc.mu.Unlock()

			if conn == nil {
				return
			}

			err := conn.WriteJSON(msg)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

// ReceiveMessages returns the receive channel
func (wc *WSClient) ReceiveMessages() <-chan WSMessage {
	return wc.receive
}
