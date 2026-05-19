package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	http.HandleFunc("/api/register", handleRegister)
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("Serwer działa na porcie 8080...")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Błąd serwera:", err)
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Rejestracja!")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Błąd podczas Upgrade:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Nowy klient połączony przez WebSocket!")

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			log.Println("Klient rozłączony lub błąd:", err)
			break
		}

		fmt.Printf("Otrzymano od klienta: %s\n", payload)

		err = conn.WriteMessage(messageType, payload)
		if err != nil {
			log.Println("Błąd podczas wysyłania odpowiedzi:", err)
			break
		}
	}
}
