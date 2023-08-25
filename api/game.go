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
	lat  float64 `json:"lat"`
	long float64 `json:"long"`
}

type message struct {
	Lat  float64 `json:"latitude"`
	Long float64 `json:"longitude"`
	Cmd  string  `json:"command"`
	Data string  `json:"data"`
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
		http.Error(w,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
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
		http.Error(w,
			http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
		return
	}

	// no gameplay should occur during the setting of the map
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// package api has global "upgrader" in socket.go
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// print error
		fmt.Printf("Game\tServeSet (/api/game/set/):\tBad connection: %v.\n", err)
		return
	}

	fmt.Print("Game\tServeSet (/api/game/set/):\tConnection opened.\n")
	conn.WriteMessage(ws.TextMessage, []byte("Please provide admin password."))

	var (
		authorized         bool
		nattempts          int
		firstCoordinate    bool
		minCoord, maxCoord coordinate
	)

	firstCoordinate = true

	// hold connection open; receive location information
	for {
		// receive message from connection
		msgType, msg, err := conn.ReadMessage()
		if err != nil || msgType == ws.CloseMessage {
			fmt.Print("Game\tServeSet (/api/game/set/):\tConnection closed")

			if err != nil {
				fmt.Printf(": %v.\n", err)
			} else {
				fmt.Print(" by user.\n")
			}

			break
		}

		var parsed message

		if err = json.Unmarshal(msg, &parsed); err != nil {
			fmt.Print("Game\tServeSet (/api/game/set/):\tCouldn't parse message: %v.\n", err)
			continue
		}

		if parsed.Cmd == "password" && parsed.Data == adminPassword {
			fmt.Print("Game\tServeSet (/api/game/set/):\tAdmin authorized.\n")
			authorized = true
			continue
		}

		if !authorized {
			fmt.Printf("Game\tServeSet (/api/game/set/):\tFailed authorization attempt; password was %q.\n", parsed.Data)
			nattempts++

			if nattempts > maxAttempts {
				fmt.Print("Game\tServeSet (/api/game/set/):\tExceeded max attempts.\n")
				conn.WriteMessage(ws.CloseMessage, []byte("Exceeded max attempts."))
				break
			}

			continue
		}

		if parsed.Cmd == "location" {
			if firstCoordinate {
				minCoord.lat = parsed.Lat
				minCoord.long = parsed.Long
				maxCoord.lat = parsed.Lat
				maxCoord.long = parsed.Long

				firstCoordinate = false
			} else {
				// find min and max values
				minCoord.lat = min(minCoord.lat, parsed.Lat)
				minCoord.long = min(minCoord.long, parsed.Long)
				maxCoord.lat = max(maxCoord.lat, parsed.Lat)
				maxCoord.long = max(maxCoord.long, parsed.Long)
			}

			fmt.Print("Game\tServeSet (/api/game/set/):\tReceived coordinates.\n")
		} else if parsed.Cmd == "write" {
			g.minCoord = minCoord
			g.maxCoord = maxCoord

			msg := fmt.Sprintf("min-lat: %f; min-long: %f; max-lat: %f; max-long: %f.",
				g.minCoord.lat, g.minCoord.long, g.maxCoord.lat, g.maxCoord.long)

			fmt.Printf("Game\tServeSet (/api/game/set/):\tUpdated min/max coordinates:\n\t%s\n", msg)

			// let admin know the set min/max coordinates
			conn.WriteMessage(ws.TextMessage, []byte(msg))
		} else {
			fmt.Printf("Game\tServeSet (/api/game/set/):\tReceived strange command: %q; ignoring.\n", parsed.Cmd)
		}
	}
}
