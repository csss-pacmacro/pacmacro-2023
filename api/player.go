// player.go

package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"strconv"
)

// user type
const (
	TypeFroshee = 0 // zero-value; froshee
	TypeLeader  = 1
	TypeAdmin   = 2
)
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

// player represents
const (
	RepsNothing = 0 // zero-value; not playing
	RepsWatcher = 1 // leaders should be this
	RepsPacman  = 2
	RepsGhost   = 3 // 3... are ghosts
	MaxGhost    = RepsGhost + 10 // permit max 10 ghosts
)
func RepsString(r uint64) string {
	switch r {
	case RepsNothing:
		return "Nothing"
	case RepsWatcher:
		return "Watcher"
	case RepsPacman:
		return "Pacman"
	default:
		return fmt.Sprintf("Ghost %d", r - RepsGhost + 1)
	}
}

const (
	// user status
	StatusGone = 0 // zero-value; out-of-game
	StatusDisc = 1 // user is disconnected; await re-connection
	StatusConn = 2 // user is connected

	id_length  = 4 // length of a session ID
)
// ID letters are hexadecimal characters
var id_letters = []rune("0123456789ABCDEF")

// zero-value player: froshee, pacman, disconnected
type Player struct {
	// private
	pass string // password

	// public
	Type   uint64 `json:"type"`
	Reps   uint64 `json:"reps"` // represents
	Status uint64 `json:"status"`
}

// attempts to log in the user with the provided password
func (p *Player) Login(pass string) bool {
	return p.pass == pass
}

func (p *Player) Format(ID string) string {
	return fmt.Sprintf("{id:%q, type:%s, reps:%s}\n", ID, TypeString(p.Type), RepsString(p.Reps))
}

type Players struct {
	adminRegistered bool
	players         map[string]*Player
	mutex           sync.Mutex
}

func (p *Players) Init() {
	p.adminRegistered = false
	p.players = make(map[string]*Player)

	fmt.Print("Players handler initialized.\n")
}

func (p *Players) New(t uint64, reps uint64, status uint64, pass string) string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var ID string

	for {
		// create random session ID
		ID_r := make([]rune, id_length)
		for i := range ID_r {
			ID_r[i] = id_letters[rand.Intn(len(id_letters))]
		}

		ID = string(ID_r)

		// break if this ID isn't in use
		if _, found := p.players[ID]; !found {
			break
		}
		// continue if found
	}

	p.players[ID] = new(Player)
	// no need to check if it was found; we just inserted it
	player, _ := p.players[ID]
	player.Type = t
	player.Reps = reps
	player.Status = status
	player.pass = pass

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

// /api/player/*
func (p *Players) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[12:]

	// GET /api/player/list.json
	if path == "list.json" {
		p.ServeList(w, r)
	// POST /api/player/register
	} else if path == "register" {
		p.ServeRegister(w, r)
	// /api/player/*
	} else {
		http.Error(w,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
	}
}

// GET /api/player/list.json
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

// POST /api/player/register
// "type": type of user
// "pass": password of user
func (p *Players) ServeRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}

	var (
		t        int
		pass, ID string
	)

	t, err := strconv.Atoi(r.FormValue("type"))
	pass = r.FormValue("pass")

	if err != nil || // invalid form data
		t < 0 || t > TypeAdmin || // validate type
		len(pass) == 0 || len(pass) > MaxPassLength { // validate pass
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}

	if t == TypeAdmin {
		// registering as admin requires the admin password
		if p.adminRegistered || pass != adminPassword {
			http.Error(w,
				http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized)
			return
		}

		// admin is registered successfully
		p.adminRegistered = true

		ID = p.New(TypeAdmin, RepsNothing, StatusDisc, pass)
	} else if t == TypeLeader {
		// register leader as a froshee watcher;
		// remains invalid until admin changes type to leader.
		ID = p.New(TypeFroshee, RepsNothing, StatusDisc, pass)
	} else {
		// register froshee into the game
		ID = p.New(TypeFroshee, RepsNothing, StatusDisc, pass)
	}

	player := p.Get(ID)
	if player == nil {
		http.Error(w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
		
	}

	fmt.Printf("Players\tServeRegister (/api/player/register/):\tRegistered ID %q as %s representing %s.\n",
		ID, TypeString(player.Type), RepsString(player.Reps))

	// respond with registered ID
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ID))
}
