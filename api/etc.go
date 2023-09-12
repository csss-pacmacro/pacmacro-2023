// etc.go
// general functions and constants that will be used in multiple files.

package api

import (
	"fmt"
	"net/http"
	ws "github.com/gorilla/websocket"
)

const (
	// general configuration
	MaxPassLength = 8      // maximum password length
	MaxAttempts   = 4      // max password attempts
	adminPassword = "1234" // NOTE: change in production

	// commands
	CMD_MOVE        = "move"        // on player movement
	CMD_INFORM      = "inform"      // inform another player change/connection

	// user type
	TypeDelete  = -1 // for /api/admin/update/<ID>
	TypeFroshee = 0 // zero-value; froshee
	TypeLeader  = 1
	TypeAdmin   = 2

	// player represents
	RepsNothing = 0 // zero-value; do not display on map
	RepsPacman  = 1
	RepsAnti    = 2 // anti-pacman; can consume pacman
	RepsGhost   = 3 // 3... are ghosts
	MaxGhost    = RepsGhost + 10 // permit max 10 ghosts

	// user status
	StatusGone = 0 // zero-value; out-of-game
	StatusDisc = 1 // user is disconnected; await re-connection
	StatusConn = 2 // user is connected

	id_length  = 4 // length of a session ID
)

var id_letters = []rune("0123456789ABCDEF")

var Upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// /*
	// FOR DEVELOPMENT: ALLOWS SERVER TO CONNECT TO ITSELF.
	// REMOVE FOR PRODUCTION
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	// */
}

// type Coordinate struct
/* JSON example:
{
	"latitude": 0,
	"longitude": 0
} */
type Coordinate struct {
	// public
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// type Message struct
/* JSON example: (NOTE: replace ... with above JSON example)
{
	"coordinate": ...,
	"command": "location",
	"data": ""
} */
type Message struct {
	// public
	Coord   Coordinate `json:"coordinate"`
	Command string     `json:"command"`
	Data    string     `json:"data"`
}

func TypeString(t uint64) string {
	switch t {
	case TypeFroshee:
		return "Froshee"
	case TypeLeader:
		return "Leader"
	case TypeAdmin:
		return "Admin"
	default:
		return "Error"
	}
}

func RepsString(r uint64) string {
	switch r {
	case RepsNothing:
		return "Nothing"
	case RepsPacman:
		return "Pacman"
	case RepsAnti:
		return "Antipac"
	default:
		return fmt.Sprintf("Ghost %d", r - RepsGhost + 1)
	}
}

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
