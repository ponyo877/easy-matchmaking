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
	port      = flag.Int("port", 8000, "The server port")
	session   = entity.NewSession[*entity.User]()
	match     = make(map[string]*entity.User)
	broadcast = make(chan *ResMsg)
)

type ReqMsg struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CloseMsg struct {
	Type string `json:"type"`
}

type ResMsg struct {
	conn      *websocket.Conn
	Type      string    `json:"type"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewResMsg(conn *websocket.Conn, roomID, userID string, createdAt time.Time) *ResMsg {
	return &ResMsg{conn, "MATCH", roomID, userID, createdAt}
}

func NewCloseMsg() *CloseMsg {
	return &CloseMsg{"CLOSE"}
}

func matchMaking() {
	for {
		if session.CanMatch() {
			now := time.Now()
			roomID := entity.NewHash(now).String()
			p1, _ := session.Dequeue()
			p2, _ := session.Dequeue()
			match[p1.ID()], match[p2.ID()] = p2, p1

			broadcast <- NewResMsg(p1.Conn(), roomID, p2.ID(), now)
			broadcast <- NewResMsg(p2.Conn(), roomID, p1.ID(), now)
			log.Printf("Matched!: %s vs %s", p1.ID(), p2.ID())
			continue
		}
	}

}

func websocketConnection(session *entity.Session[*entity.User]) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		endpoint := os.Getenv("SLACK_WEBHOOK_ENDPOINT")
		slack := notify.NewSlack(endpoint)
		_ = slack.Notify("<!here> Entry!")

		go readMessage(ws, session)
		writeMessage()
	}
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

func writeMessage() {
	for {
		res := <-broadcast
		if err := websocket.JSON.Send(res.conn, res); err != nil {
			log.Println("Error sending message to client:", err.Error())
		}
	}
}

func main() {
	flag.Parse()
	go matchMaking()
	http.HandleFunc("/matchmaking", func(w http.ResponseWriter, req *http.Request) {
		websocket.Server{Handler: websocket.Handler(websocketConnection(session))}.ServeHTTP(w, req)
	})
	log.Printf("Server listening on port %d", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatal(err)
	}
}
