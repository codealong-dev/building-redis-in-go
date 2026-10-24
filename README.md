# Building Redis in Go

> [!TIP]
> **Build this project yourself, step by step, on [Codealong](https://codealong.dev/courses/advanced-go).**
>
> This repository is the finished code. In the *Advanced Go: Building your own Redis* course you build it lesson by lesson, learning goroutines, channels, mutexes and pointers along the way. Pause the video, change the code, run it, and get instant feedback on each challenge.
>
> **[Start building your own Redis →](https://codealong.dev/courses/advanced-go)**

A small in-memory, Redis-flavoured key-value store written in Go. It was built as the final project of the [Advanced Go](https://codealong.dev/courses/advanced-go) course on Codealong and is used to practice interfaces, concurrency, the decorator (middleware) pattern, and testing.

## What's inside

```
.
├── main.go                 # Demo: replays 1000 concurrent INCRs against the store
├── command.go              # Command type, dispatch, and parallel RestoreOnBoot
├── service.go              # SetKeyWithEncryption helper (base64-encodes values)
├── middleware.go           # LoggingMiddleware: logs every Storer call
├── metrics_middleware.go   # MetricsMiddleware: call counts and latency report
└── store/
    ├── store.go            # Storer interface and shared errors
    ├── kv/                 # Thread-safe map store (sync.RWMutex) with optional max size
    └── ttl/                # Store whose keys expire after a TTL (lazy eviction)
```

### The `Storer` interface

Everything is built around one interface in `store/store.go`:

```go
type Storer interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Incr(key string) (int, error)
	Keys() []string
	Delete(key string)
	Len() int
}
```

Shared errors: `ErrEmptyKey`, `ErrStoreFull`, `ErrNotInteger`.

### Packages

- **`store/kv`**: the main engine. A `map[string]string` guarded by a `sync.RWMutex`, so many readers run in parallel while writers take an exclusive lock. `Incr` is a locked read-modify-write, so concurrent increments are never lost. `NewStore(maxSize)` caps the number of keys (`0` means unlimited) and returns `ErrStoreFull` when a new key would exceed it.
- **`store/ttl`**: a second engine where every value carries an expiry time. Expired keys are evicted lazily on `Get`, with no background goroutine. It is **not** safe for concurrent use and does not implement `Incr`, so it doesn't satisfy `Storer` yet.
- **Middlewares** (`package main`): `LoggingMiddleware` and `MetricsMiddleware` wrap any `Storer` and delegate to it. Because they implement `Storer` themselves, they can be stacked. The demo in `main.go` doesn't use them, so wrap the store yourself to try them (for example `NewLoggingMiddleware(kv.NewStore(0))`).

### Concurrent restore

`RestoreOnBoot` replays a slice of `Command`s (`SET`, `INCR`) against a `Storer`, one goroutine per command, and reports errors collected over a channel. `main.go` uses it to fire 1000 `INCR`s at one key and checks that none are lost, plus two deliberately bad commands (empty key, unknown op) to show error handling.

## Getting started

**Requirements:** Go 1.26 or newer.

```sh
git clone https://github.com/codealong-dev/building-redis-in-go.git
cd building-redis-in-go

go run .        # run the demo
go test ./...   # run the tests
go test -race ./...   # run the tests with the race detector
```

Expected output of `go run .` (the error lines may appear in either order):

```
Error: key is mandatory
Error: unknown op "WAT"
expected: 1000
got:      1000
```

## Known limitations

This is a learning project, not a production database.

- `MetricsMiddleware` counters are not synchronized, so it isn't safe to use under concurrent load. `getMisses` and the latency totals are never updated, so the `Report` output for those is always zero.
- `ttl.TTLStore` has no locking and no `Incr`. Its `Set` sleeps for one second to simulate slow writes.
- `kv.Store.Clone` copies the map without taking the lock, so don't call it while other goroutines are writing.
- Data lives in memory only. There is no persistence.

## License

[MIT](LICENSE)
