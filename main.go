package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	connections = make(map[*websocket.Conn]bool)
	mutex       = sync.Mutex{}
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	mutex.Lock()
	connections[conn] = true
	mutex.Unlock()

	for {
		// Read message from the WebSocket connection
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// Print the received message
		log.Printf("Received message: %s", message)

		// Broadcast the message back to the client
		broadcast(messageType, message)
	}

	mutex.Lock()
	delete(connections, conn)
	mutex.Unlock()
}

func broadcast(messageType int, message []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	// Iterate over all connections and send the message
	for conn := range connections {
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println("Error broadcasting message:", err)
			conn.Close()
			delete(connections, conn)
		}
	}
}

func main() {
	// Serve a basic HTML page to initiate WebSocket connection
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Handle WebSocket connections
	http.HandleFunc("/ws", handleWebSocket)

	// Start the server
	log.Println("Server starting on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
