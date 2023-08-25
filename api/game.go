package api

import (
	"sync"
	"net/http"
	"fmt"
	"encoding/json"
	ws "github.com/gorilla/websocket"
)

func min(a, b float64) float64 {
	if a < b {
		return a
	}

	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}

	return b
}

type coordinate struct {
	lat  float64 `json:"latitude"`
	long float64 `json:"longitude"`
}

type message struct {
	lat  float64 `json:"latitude"`
	long float64 `json:"longitude"`
	cmd  string  `json:"command"`
}

type Game struct {
	players *Players
	minCoord, maxCoord coordinate
	mutex sync.Mutex
}

func (g *Game) Init(players *Players) {
	g.players = players

	fmt.Print("Game handler initialized.\n")
}

func (g *Game) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[10:]

	if path == "map.json" {
		g.ServeMap(w, r)
	} else if len(path) >= 4 && path[:4] == "set/" {
		g.ServeSet(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
	}
}

func (g *Game) ServeMap(w http.ResponseWriter, r *http.Request) {
	// write map information
}

/* starts a socket connection and expects latitute/longitude values to be streamed;
 * will keep track of the minimum and maximum values received as the player walks
 * throughout the playable area. */
func (g *Game) ServeSet(w http.ResponseWriter, r *http.Request) {
	// /api/game/set/
	ID := r.URL.Path[14:] // portion after /api/game/set/

	// only admin users can set the game map
	if player := g.players.Get(ID); player == nil || player.Type != TypeAdmin {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		return
	}

	// no gameplay should occur during the setting of the map
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// package api has global "upgrader" in socket.go
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// print error
		fmt.Printf("Game ServeSet: Bad connection: %v.\n", err)
		return
	}

	var minCoord, maxCoord coordinate

	// hold connection open; receive location information
	for {
		// receive message from connection
		msgType, msg, err := conn.ReadMessage()
		if err != nil || msgType == ws.CloseMessage {
			fmt.Print("Game ServeSet: Connect closed ")

			if err != nil {
				fmt.Printf("by error: %v.\n", err)
			} else {
				fmt.Print("by user.\n")
			}

			break
		}

		var parsed message

		if json.Unmarshal(msg, parsed) != nil {
			fmt.Print("Game ServeSet: Received non-JSON message; ignoring.\n")
			continue
		}

		// find min and max values
		minCoord.lat = min(minCoord.lat, parsed.lat)
		minCoord.long = min(minCoord.long, parsed.long)
		maxCoord.lat = max(maxCoord.lat, parsed.lat)
		maxCoord.long = max(maxCoord.long, parsed.long)

		if parsed.cmd == "write" {
			g.minCoord = minCoord
			g.maxCoord = maxCoord
		}
	}
}
