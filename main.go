package main

import (
	"math/rand"
	"time"
	"net/http"
	"fmt"
	"log"

	"pacmacro/api"
)
func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var (
		players api.Players
		game    api.Game
		sock    api.Sockets
	)

	game.Init(&players) // initialize game handler
	sock.Init(&players) // initialize sockets handler

	http.Handle("/api/player/", &players) // /api/player/register; /api/player/list.json
	http.Handle("/api/game/", &game) // /api/game/map.json; /api/game/set/ID
	http.Handle("/api/ws/", &sock)

	// print to terminal that server started
	fmt.Printf("Started PacMacro; listening on localhost:8000...\n")

	// note: PacMacro API is served on port 8000 by default.
	// this should be proxied inside the web server used.
	log.Fatal(http.ListenAndServe(":8000", nil))
}
