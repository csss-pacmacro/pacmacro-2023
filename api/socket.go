// socket.go

package api

import (
	"fmt"
	"net/http"
	"sync"
	"encoding/json"
	ws "github.com/gorilla/websocket"
)

type Conn struct {
	c     *ws.Conn
	coord Coordinate
	id    string
}

type Sockets struct {
	// private
	players *Players
	conn    []*Conn // active connections; nil if broken
	mutex   sync.Mutex
}

func (s *Sockets) Init(players *Players) {
	s.players = players

	fmt.Print("Sockets handler initialized.\n")
}

// move a player; notify connections
func (s *Sockets) Move(conn_i int, coord Coordinate) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	conn := s.conn[conn_i]
	conn.coord = coord

	var msg Message
	msg.Coord   = coord
	msg.Command = CMD_MOVE
	msg.Data    = conn.id

	msg_json, err := json.Marshal(msg)
	if err != nil {
		return // failure
	}

	// for every connection
	for _, conn := range s.conn {
		// check if unactive
		if conn == nil {
			continue
		}

		// send update message
		conn.c.WriteMessage(ws.TextMessage, msg_json)
	}
}

func (s *Sockets) Connect(c *ws.Conn, id string) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	p := s.players.Get(id)
	if p == nil {
		return -1 // player doesn't exist
	}

	var conn *Conn

	conn = new(Conn)
	conn.c = c
	conn.id = id

	var msg Message
	msg.Command = CMD_INFORM
	msg.Data = p.Format(id)

	msg_json, err := json.Marshal(msg)
	if err != nil {
		fmt.Print("Err 2\n")
		return -1
	}

	conn_i := 0
	found := false

	// iterate over active connections
	for i, _ := range s.conn {
		// inform existing connections of this new connection
		if s.conn[i] != nil {
			s.conn[i].c.WriteMessage(ws.TextMessage, msg_json)
		// empty place
		} else if !found {
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

	conn_i := s.Connect(conn, ID)
	if conn_i == -1 {
		return // error
	}

	defer s.Disconnect(conn_i)

	fmt.Printf("Sockets\tServeHTTP (/api/ws/):\tID %q: Connection opened.\n", ID)

	// let player know all about themselv
	var msg_upd Message
	msg_upd.Command = CMD_UPDATE_SELF
	msg_upd.Data = player.Format(ID)
	msg_upd_json, err := json.Marshal(msg_upd)
	if err != nil {
		return // close connection; failure
	}
	err = conn.WriteMessage(ws.TextMessage, msg_upd_json)
	if err != nil {
		return // couldn't write message; failure
	}

	s.mutex.Lock()

	for i, _ := range s.conn {
		if s.conn[i] == nil {
			continue
		}

		p := s.players.Get(s.conn[i].id)
		if p == nil {
			continue
		}

		var msg Message
		msg.Coord = s.conn[i].coord
		msg.Command = CMD_INFORM
		msg.Data = p.Format(s.conn[i].id)

		msg_json, err := json.Marshal(msg)
		if err != nil {
			fmt.Print("Err 1\n")
			continue
		}

		conn.WriteMessage(ws.TextMessage, msg_json)
	}

	s.mutex.Unlock()

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
		s.Move(conn_i, coord)
	}
}
