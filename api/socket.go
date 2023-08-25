package api

import (
	"fmt"
	"net/http"
	"sync"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	// /*
	// FOR DEVELOPMENT: ALLOWS SERVER TO CONNECT TO ITSELF.
	// REMOVE FOR PRODUCTION
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	// */
}

type Sockets struct {
	players *Players
	mutex sync.Mutex
}

func (s *Sockets) Init(players *Players) {
	s.players = players

	fmt.Print("Sockets handler initialized.\n")
}

func (s *Sockets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// attempt to upgrade connection to websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// print error
		fmt.Printf("Bad connection: %v.\n", err)
		return
	}

	ID := s.players.New(TypeFroshee, RepsPacman, StatusConn, "PASSWORD")

	fmt.Printf("Sockets: ID %v: Connection opened.\n", ID)

	// hold connection open; receive location information
	for {
		// receive message from connection
		msgType, msg, err := conn.ReadMessage()
		// check if either the connection failed or was closed
		if err != nil || msgType == ws.CloseMessage {
			fmt.Printf("Sockets: ID %v: Connection closed ", ID)

			if err != nil {
				fmt.Printf("by error: %v.\n", err)
			} else {
				fmt.Print("by user.")
			}

			// exit loop; close goroutine
			break
		}

		// print received message
		fmt.Printf("Sockets: ID %v: Received message: %v.\n", ID, string(msg))

		// TODO: keep track of client's positions through messages
	}

	// disconnect this player
	s.players.SetStatus(ID, StatusDisc)
}
