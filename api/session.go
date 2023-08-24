package api

import (
	"fmt"
	"net/http"
	"sync"
)

const (
	// user type
	TypeFroshee = 0 // zero-value; froshee
	TypeLeader  = 1
	TypeAdmin   = 2

	// session represents
	RepsPacman  = 0 // zero-value; pacman
	RepsWatcher = 1 // leaders should be this
	RepsBlinky  = 2 // 2... are ghosts
	RepsPinky   = 3
	RepsInky    = 4
	RepsClyde   = 5

	// user status
	StatusDisc = 0 // zero-value; disconnected
	StatusConn = 1 // connected
)

// zero-value session: froshee, pacman, disconnected
type Session struct {
	Type   uint64 `json:"type"`
	Reps   uint64 `json:"reps"` // represents
	Status uint64 `json:"status"`
	pass   string `json:"pass"` // password
}

// attempts to log in the user with the provided password
func (s *Session) Login(_pass string) bool {
	return s.pass == _pass
}

func (s *Session) Format(ID int) string {
	return fmt.Sprintf("{id: %d; type:%d; reps:%d; status:%d}\n",
		ID, s.Type, s.Reps, s.Status)
}

type Sessions struct {
	nextID uint64
	sessions []Session
	mutex sync.Mutex
}

func (s *Sessions) New(_type uint64, _reps uint64, _status uint64, _pass string) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.sessions = append(s.sessions, Session{_type,_reps,_status,_pass})

	return len(s.sessions) - 1
}

func (s *Sessions) SetStatus(ID int, status uint64) {
	s.sessions[ID].Status = status
}

func (s *Sessions) Get(ID int) Session {
	/* notice that this is not a pointer; it is
	 * a copy of the session stored in memory. */
	return s.sessions[ID]
}

func (s *Sessions) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ret := "[\n"

	// get list of connected sessions
	for ID, session := range s.sessions {
		if session.Status == StatusConn {
			ret += session.Format(ID)
		}
	}

	ret += "]"

	w.Write([]byte(ret))
}
