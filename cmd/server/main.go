package main

import (
	"log"
	"net/http"
)

// healthHandler responds to GET /health with a small JSON body.
// In Go, an HTTP handler is just a function with this exact signature:
//   - w: where you WRITE the response (compare to Express's `res`)
//   - r: the incoming request (compare to Express's `req`)
//
// Note the reversed order vs. Express, and that you write to `w`
// directly instead of returning a value.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Set a response header. This must happen BEFORE writing the body.
	w.Header().Set("Content-Type", "application/json")

	// Set the HTTP status code. Also must come before the body.
	// http.StatusOK is a named constant for 200 (clearer than a magic number).
	w.WriteHeader(http.StatusOK)

	// Write the body. `w.Write` takes a []byte, so we convert the string.
	// The backtick string is a "raw string literal" — no escaping needed
	// for the double quotes inside the JSON.
	w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	// A ServeMux is Go's HTTP request router (like Express's Router).
	// We create our own instead of using the global default — explicit
	// is better, and it keeps our routing isolated and testable.
	mux := http.NewServeMux()

	// Register the route. Since Go 1.22, the standard library router
	// understands method + path patterns like "GET /health", so we get
	// method-based routing without any third-party package.
	mux.HandleFunc("GET /health", healthHandler)

	addr := ":8080"
	log.Printf("server listening on %s", addr)

	// ListenAndServe blocks (runs forever) until the server stops.
	// It only returns when something goes wrong, and that return value
	// is an `error`. Go has no exceptions — errors are ordinary values
	// you check explicitly. log.Fatalf prints the message and exits(1).
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
