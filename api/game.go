// game.go

package api

import (
	"fmt"
	"net/http"
)

type Game struct {
	players            *Players

	Min    Coordinate `json:"min"`
	Max    Coordinate `json:"max"`
	Width  uint64     `json:"width"`
	Height uint64     `json:"height"`
	Map    []uint64   `json:"map"`
}

func (g *Game) Init(players *Players) {
	g.players = players

	fmt.Print("Game handler initialized.\n")
}

// /api/game/*
func (g *Game) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: API calls for /api/game/*
	http.Error(w,
		http.StatusText(http.StatusNotFound),
		http.StatusNotFound)
}
