package main

import (
	"net/http"
	"log"
	"fmt"
	"pacmacro/api"
)

func main() {
	// use api.Greetings on every call to /api/
	http.HandleFunc("/api/", api.Greetings)

	// print to terminal that server started
	fmt.Printf("Started PacMacro; listening on localhost:8000...\n")

	// note: PacMacro API is served on port 8000 by default.
	// this should be proxied inside the web server used.
	log.Fatal(http.ListenAndServe(":8000", nil))
}
