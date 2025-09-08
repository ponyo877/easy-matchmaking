package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ponyo877/easy-matchmaking/entity"
	"github.com/ponyo877/easy-matchmaking/notify"
	"golang.org/x/net/websocket"
)

var (
	port    = flag.Int("port", 8000, "The server port")
	session = entity.NewSession[*entity.User]()
)

type ReqMsg struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ResMsg struct {
	Type      string    `json:"type"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewResMsg(roomID, userID string, createdAt time.Time) *ResMsg {
	return &ResMsg{"MATCH", roomID, userID, createdAt}
}

func matchmaking() {
	for {
		if session.CanMatch() {
			now := time.Now()
			roomID := entity.NewHash(now).String()
			p1, _ := session.Dequeue()
			p2, _ := session.Dequeue()

			writeMessage(p1.Conn(), NewResMsg(roomID, p2.ID(), now))
			writeMessage(p2.Conn(), NewResMsg(roomID, p1.ID(), now))
			log.Printf("Matched!: %s vs %s", p1.ID(), p2.ID())
			continue
		}
	}

}

func websocketConnection(session *entity.Session[*entity.User]) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		notifySlack()
		readMessage(ws, session)
	}
}

func notifySlack() {
	endpoint := os.Getenv("SLACK_WEBHOOK_ENDPOINT")
	slack := notify.NewSlack(endpoint)
	_ = slack.Notify("<!here> Server started!")
}

func readMessage(ws *websocket.Conn, session *entity.Session[*entity.User]) {
	mine := &entity.User{}
	for {
		var req ReqMsg
		if err := websocket.JSON.Receive(ws, &req); err != nil {
			log.Printf("Receive failed: %s; closing connection...", err.Error())
			if err = ws.Close(); err != nil {
				log.Println("Error closing connection:", err.Error())
			}
			session.Remove(mine)
			break
		}
		mine = entity.NewUser(ws, req.UserID, req.CreatedAt)
		session.Add(mine)
		log.Printf("New entry: %s, from %s\n", req.UserID, ws.Request().RemoteAddr)
	}
}

func writeMessage(ws *websocket.Conn, res *ResMsg) {
	if err := websocket.JSON.Send(ws, res); err != nil {
		log.Println("Error sending message to client:", err.Error())
	}
}

func main() {
	flag.Parse()
	go matchmaking()
	http.HandleFunc("/matchmaking", func(w http.ResponseWriter, req *http.Request) {
		websocket.Server{Handler: websocket.Handler(websocketConnection(session))}.ServeHTTP(w, req)
	})
	log.Printf("Server listening on port %d", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatal(err)
	}
}
