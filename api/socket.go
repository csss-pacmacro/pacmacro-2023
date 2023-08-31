// socket.go

package api

import (
	"fmt"
	"net/http"
	"sync"
	"encoding/json"
	ws "github.com/gorilla/websocket"
)

type Sockets struct {
	// private
	players *Players
	coord   map[string]Coordinate
	conn    []*ws.Conn // active connections; nil if broken
	mutex   sync.Mutex
}

func (s *Sockets) Init(players *Players) {
	s.players = players
	s.coord = make(map[string]Coordinate)

	fmt.Print("Sockets handler initialized.\n")
}

/* JSON example:
{
	"id": "ABCD",
	"coordinate": ...
} */
type PlayerUpdate struct {
	ID    string     `json:"id"`
	Coord Coordinate `json:"coordinate"`
}

// move a player; notify connections
func (s *Sockets) Move(ID string, coord Coordinate) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// make JSON update message
	upd := PlayerUpdate{ID, coord}

	msg, err := json.Marshal(upd)
	if err != nil {
		return // shouldn't occur
	}

	s.coord[ID] = coord

	// for every connection
	for _, conn := range s.conn {
		// check if unactive
		if conn == nil {
			continue
		}

		// send update message
		conn.WriteMessage(ws.TextMessage, msg)
	}
}

func (s *Sockets) Connect(conn *ws.Conn) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	conn_i := 0
	found := false

	// iterate over active connections
	for i, _ := range s.conn {
		// empty place
		if s.conn[i] == nil {
			s.conn[i] = conn
			conn_i = i
			found = true
		}
	}

	// if there are no empty spaces; append
	if !found {
		conn_i = len(s.conn)
		s.conn = append(s.conn, conn)
	}

	return conn_i
}

func (s *Sockets) Disconnect(conn_i int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.conn[conn_i] = nil
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

	// attempt to log in
	msgType, msg, err := conn.ReadMessage()
	if err != nil || msgType != ws.TextMessage || !player.Login(string(msg)) {
		conn.Close()
		return
	}

	conn_i := s.Connect(conn)
	defer s.Disconnect(conn_i)

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

		var coord Coordinate

		if err := json.Unmarshal(msg, &coord); err != nil {
			fmt.Printf("Sockets\tServeHTTP (/api/ws/):\tID %q: Received invalid message: %s.\n", ID, string(msg))
			continue
		}

		// notify other connections of movement
		s.Move(ID, coord)
	}
}
