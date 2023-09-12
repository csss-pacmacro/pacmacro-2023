// game.go

package api

import (
	"fmt"
	"net/http"
	"sync"
	"encoding/json"
)

type Game struct {
	// private
	players *Players
	mutex   sync.Mutex

	// public
	Min    Coordinate `json:"min"`
	Max    Coordinate `json:"max"`
	Width  uint64     `json:"width"`
	Height uint64     `json:"height"`
}

func (g *Game) Init(players *Players) {
	g.players = players

	// arbitrary default values; for testing
	g.Width = 32
	g.Height = 32

	fmt.Print("Game handler initialized.\n")
}

// /api/game/*
func (g *Game) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[10:]

	// GET /api/game/map.json
	if path == "map.json" {
		g.ServeMap(w, r)
	// /api/game/*
	} else {
		http.Error(w,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
	}
}

// GET /api/game/map.json
func (g *Game) ServeMap(w http.ResponseWriter, r *http.Request) {
	JSON, err := json.Marshal(g)
	if err != nil {
		http.Error(w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Write(JSON)
}
