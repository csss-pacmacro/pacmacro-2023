package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
)

const (
	// user type
	TypeFroshee = 0 // zero-value; froshee
	TypeLeader  = 1
	TypeAdmin   = 2

	// player represents
	RepsPacman  = 0 // zero-value; pacman
	RepsWatcher = 1 // leaders should be this
	RepsBlinky  = 2 // 2... are ghosts
	RepsPinky   = 3
	RepsInky    = 4
	RepsClyde   = 5

	// user status
	StatusGone = 0 // zero-value; out-of-game
	StatusDisc = 1 // user is disconnected; await re-connection
	StatusConn = 2 // user is connected

	id_length  = 4
)

var id_letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// zero-value player: froshee, pacman, disconnected
type Player struct {
	Type   uint64 `json:"type"`
	Reps   uint64 `json:"reps"` // represents
	Status uint64 `json:"status"`
	pass   string `json:"pass"` // password
}

// attempts to log in the user with the provided password
func (p *Player) Login(_pass string) bool {
	return p.pass == _pass
}

func (p *Player) Format(ID string) string {
	return fmt.Sprintf("{id: %d, type:%d, reps:%d}\n", ID, p.Type, p.Reps)
}

type Players struct {
	players map[string]*Player
	mutex sync.Mutex
}

func (p *Players) New(_type uint64, _reps uint64, _status uint64, _pass string) string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// create random session ID
	ID_r := make([]rune, id_length)
	for i := range ID_r {
		ID_r[i] = id_letters[rand.Intn(len(id_letters))]
	}

	ID := string(ID_r)

	p.players[ID] = new(Player)
	// no need to check if it was found; we just inserted it
	player, _ := p.players[ID]
	player.Type = _type
	player.Reps = _reps
	player.Status = _status
	player.pass = _pass

	return ID
}

func (p *Players) Delete(ID string) {
	delete(p.players, ID)
}

func (p *Players) SetStatus(ID string, status uint64) {
	if player, found := p.players[ID]; found {
		player.Status = status
	}
}

func (p *Players) Get(ID string) *Player {
	if player, found := p.players[ID]; found {
		return player
	} else {
		return nil
	}
}

// /api/player/
func (p *Players) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[12:]

	if path == "list.json" {
		p.ServeList(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
	}
}

func (p *Players) ServeList(w http.ResponseWriter, r *http.Request) {
	ret := "[\n"

	// get list of connected players
	for ID, player := range p.players {
		if player.Status != StatusGone {
			ret += player.Format(ID)
		}
	}

	ret += "]"

	w.Write([]byte(ret))
}
