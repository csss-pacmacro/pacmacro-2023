package api

import (
	"fmt"
	"net/http"
)

func Greetings(w http.ResponseWriter, r *http.Request) {
	// format path into a greetings message
	msg := fmt.Sprintf("Greetings! You accessed: %s.", r.URL.Path)

	// write message as bytes
	w.Write([]byte(msg))
}
