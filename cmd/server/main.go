package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// shortenRequest is the shape of the JSON we expect in the POST body.
type shortenRequest struct {
	URL string `json:"url"` // The `json:"url"` tag tells the JSON decoder to look for "url" in the input.
}

// shortenResponse is the shape of the JSON we send back in the response.
type shortenResponse struct {
	Code string `json:"code"` // The `json:"code"` tag tells the JSON encoder to use "code" as the key in the output.
}

// statsResponse is the shape of the JSON we send back from GET /{code}/stats.
type statsResponse struct {
	Count int `json:"count"` // how many times this code has been redirected
}

type entry struct {
	URL   string
	Count int
}

type Store interface {
	save(url string) string
	get(code string) (string, bool)
	stats(code string) (int, bool)
}

// Compile-time check that *store satisfies Store. If a method signature ever
// drifts, the build fails here — at the type — instead of at a call site.
var _ Store = (*store)(nil)

// store holds the code→URL mapping. The mutex guards the map so that
// concurrent requests (each in its own goroutine) can't corrupt it.
type store struct {
	mu   sync.Mutex
	urls map[string]*entry // maps code to entry (URL + Count together)
}

func newStore() *store {
	return &store{
		urls: make(map[string]*entry),
	}
}

// save records a url under a freshly generated code and returns that code.
func (s *store) save(url string) string {
	s.mu.Lock()
	defer s.mu.Unlock() // runs when the function returns -- releases the lock
	code := generateCode()

	s.urls[code] = &entry{URL: url} // build a *entry; Count starts at 0

	return code
}

// get returns the URL for a code and counts the lookup as one click.
func (s *store) get(code string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.urls[code]
	if !ok {
		return "", false
	}
	e.Count++ // mutate through the pointer — this is why the map holds *entry
	return e.URL, true
}

// stats returns the click count for a code WITHOUT incrementing it —
// reading stats is not itself a click.
func (s *store) stats(code string) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.urls[code]
	if !ok {
		return 0, false
	}
	return e.Count, true
}

func generateCode() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func shortenHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		req := &shortenRequest{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if req.URL == "" {
			http.Error(w, "url is required", http.StatusBadRequest)
			return
		}

		code := s.save(req.URL)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(shortenResponse{Code: code})
	}
}

func redirectHandler(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code") // extract the {code} part from the URL
		url, ok := s.get(code)

		if !ok {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
	}
}

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

	// endpoint list:

	// GET /health → healthHandler
	// Register the route. Since Go 1.22, the standard library router
	// understands method + path patterns like "GET /health", so we get
	// method-based routing without any third-party package.
	mux.HandleFunc("GET /health", healthHandler)

	// POST /shorten → shortenHandler
	// Register the handler for POST /shorten.
	s := newStore() // create a new store to hold our URL mappings
	mux.HandleFunc("POST /shorten", shortenHandler(s))

	// GET /{code} → redirectHandler
	mux.HandleFunc("GET /{code}", redirectHandler(s))

	// GET /{code}/stats → statsHandler
	mux.HandleFunc("GET /{code}/stats", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		count, ok := s.stats(code) // stats, not get — reading must not count as a click
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statsResponse{Count: count})
	})

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
