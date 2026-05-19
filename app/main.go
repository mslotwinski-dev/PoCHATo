package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("Uruchamianie klienta poCHATo...")
	wsURL := "ws://localhost:8080/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Błąd połączenia WebSocket:", err)
	}
	defer conn.Close()

	msg := []byte("Cześć serwerze! To jest tajna wiadomość E2EE (jeszcze nie!).")
	err = conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Fatal("Błąd wysyłania wiadomości:", err)
	}
	fmt.Println("Wysłano wiadomość do serwera.")

	_, response, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("Błąd odbierania odpowiedzi:", err)
	}

	fmt.Printf("Otrzymano odpowiedź: %s\n", response)
}
