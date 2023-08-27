// game.go

package api

import (
	"fmt"
	"net/http"
)

type Game struct {
	players            *Players
	minCoord, maxCoord Coordinate
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
