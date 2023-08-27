// socket.go

package api

import (
	"fmt"
	"net/http"
	"sync"
	ws "github.com/gorilla/websocket"
)

type Sockets struct {
	players *Players
	mutex   sync.Mutex
}

func (s *Sockets) Init(players *Players) {
	s.players = players

	fmt.Print("Sockets handler initialized.\n")
}

// WS /api/ws/<ID>
func (s *Sockets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ID := r.URL.Path[8:]
	player := s.players.Get(ID)

	if player == nil {
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		// print error
		fmt.Printf("Sockets\tServeHTTP (/api/ws/):\tBad Request; ID %q not registered.\n", ID)
		return
	}

	// attempt to upgrade connection to websocket connection
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		// print error
		fmt.Printf("Sockets\tServeHTTP (/api/ws/):\tBad connection: %v.\n", err)
		return
	}

	fmt.Printf("Sockets\tServeHTTP (/api/ws/):\tID %q: Connection opened.\n", ID)

	if err = conn.WriteMessage(ws.TextMessage, []byte("Hello from the API!")); err != nil {
		fmt.Printf("Sockets\tServeHTTP (/api/ws/):\tCouldn't write message to %q.\n", ID)
		// attempt to close connection
		conn.WriteMessage(ws.CloseMessage, []byte("Goodbye"))
		return
	}

	s.players.SetStatus(ID, StatusConn)
	defer s.players.SetStatus(ID, StatusDisc)

	// hold connection open; receive location information
	for {
		// receive message from connection
		msgType, msg, err := conn.ReadMessage()
		// check if either the connection failed or was closed
		if err != nil || msgType == ws.CloseMessage {
			fmt.Printf("Sockets\tServeHTTP (/api/ws/):\tID %q: Connection closed", ID)

			if err != nil {
				fmt.Printf(": %v.\n", err)
			} else {
				fmt.Print(" by user.")
			}

			// exit loop; close goroutine
			break
		}

		// print received message
		fmt.Printf("Sockets\tServeHTTP (/api/ws/):\tID %q: Received message: %v.\n", ID, string(msg))

		// TODO: keep track of client's positions through messages
	}
}
