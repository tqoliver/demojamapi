package main

import (
	"fmt"
	"log"
	"net/http"
)

// helloHandler handles requests to the root path "/"
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header to indicate plain text
	w.Header().Set("Content-Type", "text/plain")

	// Write the "Hello, World!" message to the response writer
	fmt.Fprint(w, "Hello, World!")
}

func main() {
	// Register the helloHandler function to handle requests to the root path
	http.HandleFunc("/", helloHandler)

	// Start the HTTP server on port 8080
	// log.Fatal will print an error and exit if the server fails to start
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
