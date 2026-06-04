# go-url-shortener

A small URL shortener written in Go, using only the standard library. It
exposes an HTTP API to shorten a URL into a short code and to redirect from
that code back to the original URL.

This is a learning project, built incrementally to explore Go's `net/http`
server, the standard-library router, JSON handling, and concurrency-safe
state.

## Requirements

- Go 1.26 or newer

## Running

```bash
go run ./cmd/server
```

The server listens on `:8080`:

```
server listening on :8080
```

To build a binary instead:

```bash
go build -o server ./cmd/server
./server
```

## API

### `GET /health`

Liveness check.

```bash
curl localhost:8080/health
```

```json
{"status":"ok"}
```

### `POST /shorten`

Store a URL and get back a short code. The body is JSON with a `url` field.

```bash
curl -X POST localhost:8080/shorten -d '{"url":"https://go.dev"}'
```

```json
{"code":"dizz6k68c63a"}
```

Returns `400 Bad Request` if the body isn't valid JSON or if `url` is empty.

### `GET /{code}`

Redirect to the original URL for a code.

```bash
curl -i localhost:8080/dizz6k68c63a
```

```
HTTP/1.1 302 Found
Location: https://go.dev
```

Returns `404 Not Found` if the code isn't known.

## Project layout

```
cmd/server/main.go   # server entrypoint, store, and HTTP handlers
go.mod               # module definition
```

## Notes and limitations

- **Storage is in-memory.** All shortened URLs are held in a map and are lost
  when the server stops.
- Short codes are derived from a nanosecond timestamp, so they are unique per
  run but not random or guessable-resistant.

## License

See [LICENSE](LICENSE).
