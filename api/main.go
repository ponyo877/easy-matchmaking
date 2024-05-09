package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

var (
	port      = flag.Int("port", 8000, "The server port")
	sessions  = make(map[*websocket.Conn]bool)
	broadcast = make(chan Message)
)

type Message struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

func NewMessage(text string, createdAt time.Time) Message {
	return Message{text, createdAt}
}

func websocketEchoConnection(ws *websocket.Conn) {
	for {
		sessions[ws] = true
		var message Message
		if err := websocket.JSON.Receive(ws, &message); err != nil {
			log.Printf("Receive failed: %s; closing connection...", err.Error())
			if err = ws.Close(); err != nil {
				log.Println("Error closing connection:", err.Error())
			}
		}
		broadcast <- message
	}
}

func writeMessage() {
	for {
		message := <-broadcast
		for session := range sessions {
			if err := websocket.JSON.Send(session, message); err != nil {
				log.Println("Error sending message to client:", err.Error())
			}
		}
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		websocket.Server{Handler: websocket.Handler(websocketEchoConnection)}.ServeHTTP(w, req)
	})
	log.Printf("Server listening on port %d", *port)

	go writeMessage()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
