// etc.go
// general functions and constants that will be used in multiple files.

package api

import (
	"net/http"
	ws "github.com/gorilla/websocket"
)

const (
	MaxPassLength = 8      // maximum password length
	MaxAttempts   = 4      // max password attempts
	adminPassword = "1234" // NOTE: change in production
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
