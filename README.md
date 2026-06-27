# todo-api — TODO REST API in Go

> HTTP REST API for task management with an in-memory store. Learning project #4 in a Go Backend Developer preparation roadmap.

---

## For Recruiters

### What this is and why it exists

The fourth project in the roadmap — same business logic as todo-cli, but now exposed over HTTP. This is a classic transition from a CLI tool to a network service: `os.Args` becomes JSON requests, a file becomes an in-memory store protected against race conditions via `sync.RWMutex`.

The main goal is to learn building a full CRUD REST API using only `net/http`: correct routing with methods and path parameters, a unified response format, REST-conventional HTTP status codes, and a thread-safe store. This mirrors exactly how internal microservices are built at Ozon and WB — Go + `net/http` + in-memory state or an external DB.

The project is intentionally framework-free and without persistence — the focus is on the correct HTTP layer architecture and concurrent data access.

### What this project demonstrates

| Skill | Implementation |
|---|---|
| REST CRUD | `POST /todos`, `GET /todos`, `GET /todos/{id}`, `PATCH /todos/{id}`, `DELETE /todos/{id}` |
| Routing with path params | `http.NewServeMux` + Go 1.22 patterns: `GET /todos/{id}` |
| Thread safety | `sync.RWMutex` — `RLock` for reads, `Lock` for writes |
| In-memory store | `Store` struct encapsulates `map[int]Todo` + mutex |
| JSON API | unified response format, `Content-Type: application/json` |
| HTTP status codes | 200, 201, 204, 400, 404, 405 per REST convention |
| Middleware | logging middleware: method + path + status + duration |
| Layer separation | `store.go` (data) + `handler.go` (HTTP) + `main.go` (wiring) |

### Stack

- **Language:** Go 1.22+
- **Dependencies:** standard library only
- **Store:** in-memory (`map[int]Todo` + `sync.RWMutex`)
- **Platform:** Linux / macOS / Windows

---

## For Developers

### Architecture decisions

#### Why in-memory and not a JSON file like in #3?

The file from project #3 is not thread-safe — two concurrent requests can read and write simultaneously. An in-memory store with `sync.RWMutex` solves this cleanly. Persistence is a task for project #7 (PostgreSQL). The goal here is the correct concurrency model.

#### Why `sync.RWMutex` and not `sync.Mutex`?

`RWMutex` allows multiple goroutines to read concurrently (`RLock`) and blocks everyone during writes (`Lock`). For read-heavy workloads (many GETs, few POSTs/DELETEs) this is a meaningful performance gain. `Mutex` blocks even parallel reads — unnecessarily restrictive.

#### Why `Store` as a struct with methods and not a global variable?

```go
// Bad — global state, cannot be tested in isolation
var todos = map[int]Todo{}
var mu sync.RWMutex

// Good — encapsulation, can create multiple instances for tests
type Store struct {
    mu    sync.RWMutex
    todos map[int]Todo
    next  int
}
```

Encapsulating the mutex inside Store guarantees that no external code can access the map without holding the lock.

#### Why `PATCH /todos/{id}` and not `PUT`?

`PUT` is a full resource replacement — the client must send all fields. `PATCH` is a partial update. For the "mark as done" operation (`done: true`), the correct method is `PATCH`: only the changed field is sent.

#### Why 201 on creation and not 200?

REST convention: `200 OK` — request succeeded, `201 Created` — resource was created. The distinction matters for clients that process API responses automatically. A `Location` header with the URL of the created resource is a bonus that reviewers appreciate.

### Structure

```
todo-api/
├── main.go       # wiring: NewStore, handler registration, ListenAndServe
├── store.go      # Store struct + methods: GetAll, GetByID, Create, Update, Delete
├── handler.go    # HTTP handlers: decode request, call store, encode response
├── middleware.go # loggingMiddleware: method + path + status + duration
├── go.mod
└── README.md
```

### Installation and running

```bash
git clone https://github.com/Shipovmax/todo-api
cd todo-api
go run .
# Server started on :8080
```

### API

| Method | Path | Description | Status |
|--------|------|-------------|--------|
| `POST` | `/todos` | Create a task | 201 |
| `GET` | `/todos` | Get all tasks | 200 |
| `GET` | `/todos/{id}` | Get task by ID | 200 |
| `PATCH` | `/todos/{id}` | Update task (done) | 200 |
| `DELETE` | `/todos/{id}` | Delete task | 204 |

### Examples

```bash
# Create a task
curl -s -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn sync.RWMutex"}'
# HTTP 201: {"id":1,"title":"Learn sync.RWMutex","done":false,"created_at":"..."}

# Get all
curl -s http://localhost:8080/todos
# HTTP 200: [{"id":1,"title":"Learn sync.RWMutex","done":false,...}]

# Get by ID
curl -s http://localhost:8080/todos/1
# HTTP 200: {"id":1,"title":"Learn sync.RWMutex","done":false,...}

# Mark as done
curl -s -X PATCH http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"done": true}'
# HTTP 200: {"id":1,"title":"Learn sync.RWMutex","done":true,...}

# Delete
curl -s -X DELETE http://localhost:8080/todos/1
# HTTP 204 (empty body)
```

### Error handling

```bash
# Task not found
curl -s http://localhost:8080/todos/99
# HTTP 404: {"error":"task not found"}

# Invalid JSON
curl -s -X POST http://localhost:8080/todos -d 'not json'
# HTTP 400: {"error":"invalid JSON"}

# Empty title
curl -s -X POST http://localhost:8080/todos -d '{"title":""}'
# HTTP 400: {"error":"title cannot be empty"}

# Invalid ID in path
curl -s http://localhost:8080/todos/abc
# HTTP 400: {"error":"invalid ID"}

# Wrong method
curl -s -X PUT http://localhost:8080/todos
# HTTP 405
```

### Running without building

```bash
go run .
```
