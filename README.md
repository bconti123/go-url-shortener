# go-url-shortener

A small URL shortener written in Go. It exposes an HTTP API to shorten a URL
into a short code and to redirect from that code back to the original URL.
Each short code also tracks how many times it has been visited.

This is a learning project, built incrementally to explore Go's `net/http`
server, the standard-library router, JSON handling, concurrency-safe state,
interfaces, and `database/sql`.

## Requirements

- Go 1.26 or newer
- Docker (optional, only for the Postgres backend)

## Running

By default the server uses an in-memory store and needs nothing else:

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

### With Postgres

Set `DATABASE_URL` and the server uses Postgres instead of the in-memory
store. A `docker-compose.yml` is included to run Postgres locally (published
on host port `5440`, since `5432`–`5434` are often taken by local clusters):

```bash
docker compose up -d
export DATABASE_URL='postgres://urlshort:urlshort@localhost:5440/urlshort?sslmode=disable'
go run ./cmd/server
```

On startup the server logs which backend it chose (`using Postgres store` or
`using in-memory store`).

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

Each redirect increments a per-code click counter (see `GET /{code}/stats`).

### `GET /{code}/stats`

Report how many times a code has been redirected.

```bash
curl localhost:8080/dizz6k68c63a/stats
```

```json
{"count":3}
```

Reading stats does **not** count as a click. Returns `404 Not Found` if the
code isn't known.

## Project layout

```
cmd/server/main.go       # entrypoint, HTTP handlers, Store interface, in-memory store
cmd/server/postgres.go   # Postgres connection and Store implementation
cmd/server/*_test.go     # store unit tests and a gated Postgres integration test
db/schema.sql            # urls table, run on first container init
docker-compose.yml       # local Postgres
go.mod                   # module definition
```

Both stores satisfy a common `Store` interface, so the HTTP handlers are
unaware of which backend is in use; `main` picks one based on `DATABASE_URL`.

## Notes and limitations

- **The in-memory store is not persistent.** Without `DATABASE_URL`, all
  shortened URLs and their click counts live in a map and are lost when the
  server stops. The Postgres backend persists across restarts.
- Short codes are derived from a nanosecond timestamp, so they are unique per
  run but not random or guessable-resistant.

## License

See [LICENSE](LICENSE).
