// admin.go

package api

import (
	"sync"
	"net/http"
	"fmt"
	"encoding/json"
	"strconv"
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

	// POST /api/admin/scale
	if path == "scale" {
		a.ServeScale(w, r)
	// POST /api/admin/populate
	} else if path == "populate" {
		a.ServePopulate(w, r)
	// WS /api/admin/set/<ID>
	} else if len(path) >= 4 && path[:4] == "set/" {
		a.ServeSet(w, r)
	// /api/admin/*
	} else {
		http.Error(w,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
	}
}

// POST /api/admin/scale
// "id": admin ID
// "pass": admin password
// "width": horizontal scale of map
// "height": vertical scale of map
func (a *Admin) ServeScale(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}

	ID := r.FormValue("id")
	pass := r.FormValue("pass")

	// authorize player as admin
	if p := a.players.Get(ID); (p != nil && p.Type != TypeAdmin) || pass != adminPassword {
		http.Error(w,
			http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
		return
	}

	var wi, h int

	wi, err := strconv.Atoi(r.FormValue("width"))
	if err != nil {
		h, err = strconv.Atoi(r.FormValue("height"))
	}
	pass = r.FormValue("pass")

	// ensure that submitted form data is valid
	if err != nil || wi < 0 || h < 0 {
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}

	a.game.Width = uint64(wi)
	a.game.Height = uint64(h)
}

// POST /api/admin/populate
// "id": admin ID
// "pass": admin password
// "map": JSON array storing map information
func (a *Admin) ServePopulate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}

	ID := r.FormValue("id")
	pass := r.FormValue("pass")

	// authorize player as admin
	if p := a.players.Get(ID); (p != nil && p.Type != TypeAdmin) || pass != adminPassword {
		http.Error(w,
			http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
		return
	}

	var (
		map_s string
		// NOTE: json.Marshal interprets all JSON numbers as float64;
		// thus, this preliminary map stores each value as a float64
		m []float64
	)

	// parse JSON map
	map_s = r.FormValue("map")
	if err := json.Unmarshal([]byte(map_s), &m); err != nil {
		fmt.Printf("Admin\tServePopulate (/api/admin/populate)\tFailure parsing map data: %v;\n" +
			"----BEGIN MAP JSON\n%s\n----END MAP JSON\n", err, map_s)
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}

	// create a new map
	a.game.Map = make([]uint64, len(m))

	// cast every float64 from m to a uint64 in a.game.Map
	for i, f := range m {
		a.game.Map[i] = uint64(f)
	}

	fmt.Printf("Admin\tServePopulate (/api/admin/populate)\tPopulated map;\n" +
		"----BEGIN MAP DATA\n%+v\n----END MAP DATA", a.game.Map)
}

// WS /api/admin/set/<ID>
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
		authorized   bool
		nattempts    int
		first        bool
		min_c, max_c Coordinate
	)

	// first coordinate should immediately set min/max values
	first = true

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
			if first {
				min_c = parsed.Coord
				max_c = parsed.Coord

				first = false
			} else {
				// find min and max values
				min_c.Latitude = min(min_c.Latitude, parsed.Coord.Latitude)
				min_c.Longitude = min(min_c.Longitude, parsed.Coord.Longitude)
				max_c.Latitude = max(max_c.Latitude, parsed.Coord.Latitude)
				max_c.Longitude = max(max_c.Longitude, parsed.Coord.Longitude)
			}

			fmt.Print("Admin\tServeSet (/api/admin/set/):\tReceived coordinates.\n")
			conn.WriteMessage(ws.TextMessage, []byte("Received coordinates."))
		} else if parsed.Command == "write" {
			a.game.Min = min_c
			a.game.Max = max_c

			msg := fmt.Sprintf("Min: %+v\nMax: %+v", a.game.Min, a.game.Max)
			fmt.Printf("Admin\tServeSet (/api/admin/set/):\tUpdated game.Min and game.Max;\n" +
				"----BEGIN SUMMARY\n%s\n----END SUMMARY\n", msg)

			// let admin know the set min/max coordinates
			conn.WriteMessage(ws.TextMessage, []byte(msg))
		} else {
			fmt.Printf("Admin\tServeSet (/api/admin/set/):\tReceived strange command: %q; ignoring.\n", parsed.Command)
		}
	}
}
