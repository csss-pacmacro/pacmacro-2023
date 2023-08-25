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

	// length of a session ID
	id_length  = 4

	// maximum password length
	MaxPassLength = 8

	// password for admin to register with; CHANGE IN PRODUCTION
	// NOTE: only the admin can make someone a leader
	adminPassword = "1234"
)

// ID letters are hexadecimal characters
var id_letters = []rune("0123456789ABCDEF")

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

func (p *Players) New(_type uint64, _reps uint64, _status uint64, _pass string) string {
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
	} else if path == "register" {
		p.ServeRegister(w, r)
	} else {
		http.Error(w,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
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
		_type     int
		_pass, ID string
	)

	_type, err := strconv.Atoi(r.FormValue("type"))
	_pass = r.FormValue("pass")

	if err != nil || // invalid form data
		_type < 0 || _type > TypeAdmin || // validate type
		len(_pass) == 0 || len(_pass) > MaxPassLength { // validate pass
		http.Error(w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}

	if _type == TypeAdmin {
		// registering as admin requires the admin password
		if p.adminRegistered || _pass != adminPassword {
			http.Error(w,
				http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized)
			return
		}

		// admin is registered successfully
		p.adminRegistered = true

		ID = p.New(TypeAdmin, RepsNothing, StatusDisc, _pass)
	} else if _type == TypeLeader {
		// register leader as a froshee watcher;
		// remains invalid until admin changes type to leader.
		ID = p.New(TypeFroshee, RepsWatcher, StatusDisc, _pass)
	} else {
		// register froshee into the game
		ID = p.New(TypeFroshee, RepsNothing, StatusDisc, _pass)
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
