package main

import (
	"log"
	"net/http"
)

func main() {
	// Serve files from current directory (cmd/test-client)
	// When running from root: go run cmd/test-client/main.go -> we need to point to cmd/test-client
	// But if we run it, we might need to adjust path.
	// Easiest: http.Dir(".") if we run from cmd/test-client?
	// Or hardcode logic.

	dir := "./cmd/test-client"

	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	log.Println("---------------------------------------------------------")
	log.Println("   Test Client running at http://localhost:3000")
	log.Println("   Open this URL in your browser")
	log.Println("---------------------------------------------------------")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
