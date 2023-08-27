// admin.go

package api

import (
	"sync"
	"net/http"
	"fmt"
	"encoding/json"
	ws "github.com/gorilla/websocket"
)

type Admin struct {
	game    *Game
	players *Players
	mutex   sync.Mutex
}

func (a *Admin) Init(game *Game, players *Players) {
	a.game = game
	a.players = players

	fmt.Print("Admin handler initialized.\n")
}

// /api/admin/*
func (a *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[11:]

	// /api/admin/set/<ID>
	if len(path) >= 4 && path[:4] == "set/" {
		a.ServeSet(w, r)
	// /api/admin/*
	} else {
		http.Error(w,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
	}
}

// /api/admin/set/<ID>
/* starts a socket connection and expects latitute/longitude values to be streamed;
 * will keep track of the minimum and maximum values received as the player walks
 * throughout the playable area. */
func (a *Admin) ServeSet(w http.ResponseWriter, r *http.Request) {
	// /api/admin/set/
	ID := r.URL.Path[15:] // portion after /api/admin/set/

	// only admin users can set the game map
	if player := a.players.Get(ID); player == nil || player.Type != TypeAdmin {
		http.Error(w,
			http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
		return
	}

	// no gameplay should occur during the setting of the map
	a.mutex.Lock()
	defer a.mutex.Unlock()

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		// print error
		fmt.Printf("Admin\tServeSet (/api/admin/set/):\tBad connection: %v.\n", err)
		return
	}

	fmt.Print("Admin\tServeSet (/api/admin/set/):\tConnection opened.\n")
	conn.WriteMessage(ws.TextMessage, []byte("Waiting for password authentication..."))

	var (
		authorized         bool
		nattempts          int
		firstCoordinate    bool
		minCoord, maxCoord Coordinate
	)

	// first coordinate should immediately set min/max values
	firstCoordinate = true

	// hold connection open; receive location information
	for {
		// receive message from connection
		msgType, msg, err := conn.ReadMessage()
		if err != nil || msgType == ws.CloseMessage {
			fmt.Print("Admin\tServeSet (/api/admin/set/):\tConnection closed")

			if err != nil {
				fmt.Printf(": %v.\n", err)
			} else {
				fmt.Print(" by user.\n")
			}

			break
		}

		var parsed Message

		if err = json.Unmarshal(msg, &parsed); err != nil {
			fmt.Printf("Admin\tServeSet (/api/admin/set/):\tCouldn't parse message: %v;\n" +
				"----BEGIN MESSAGE\n%s\n----END MESSAGE\n", err, string(msg))
			continue
		}

		if parsed.Command == "password" && parsed.Data == adminPassword {
			fmt.Print("Admin\tServeSet (/api/admin/set/):\tAdmin authorized.\n")
			conn.WriteMessage(ws.TextMessage, []byte("Authorized."))
			authorized = true
			continue
		}

		if !authorized {
			fmt.Printf("Admin\tServeSet (/api/admin/set/):\tFailed authorization attempt; password was %q.\n", parsed.Data)
			nattempts++

			if nattempts > MaxAttempts {
				fmt.Print("Admin\tServeSet (/api/admin/set/):\tExceeded max attempts.\n")
				conn.WriteMessage(ws.CloseMessage, []byte("Exceeded max attempts."))
				break
			}

			continue
		}

		if parsed.Command == "location" {
			if firstCoordinate {
				minCoord = parsed.Coord
				maxCoord = parsed.Coord

				firstCoordinate = false
			} else {
				// find min and max values
				minCoord.Latitude = min(minCoord.Latitude, parsed.Coord.Latitude)
				minCoord.Longitude = min(minCoord.Longitude, parsed.Coord.Longitude)
				maxCoord.Latitude = max(maxCoord.Latitude, parsed.Coord.Latitude)
				maxCoord.Longitude = max(maxCoord.Longitude, parsed.Coord.Longitude)
			}

			fmt.Print("Admin\tServeSet (/api/admin/set/):\tReceived coordinates.\n")
			conn.WriteMessage(ws.TextMessage, []byte("Received coordinates."))
		} else if parsed.Command == "write" {
			a.game.minCoord = minCoord
			a.game.maxCoord = maxCoord

			msg := fmt.Sprintf("minCoord: %+v\nmaxCoord: %+v", a.game.minCoord, a.game.maxCoord)
			fmt.Printf("Admin\tServeSet (/api/admin/set/):\tUpdated game.minCoord and game.maxCoord;\n" +
				"----BEGIN SUMMARY\n%s\n----END SUMMARY\n", msg)

			// let admin know the set min/max coordinates
			conn.WriteMessage(ws.TextMessage, []byte(msg))
		} else {
			fmt.Printf("Admin\tServeSet (/api/admin/set/):\tReceived strange command: %q; ignoring.\n", parsed.Command)
		}
	}
}
