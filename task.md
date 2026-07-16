# Task #4 — TODO HTTP API

## Goal

Write a REST API for task management using plain `net/http` with an in-memory store. The main learning goal is to master a full CRUD flow over HTTP, Go 1.22+ path-parameter routing, and thread-safe access to shared data via `sync.RWMutex`. Concurrent access to shared state is one of the top topics in Go interviews at Ozon/WB/Sberbank.

---

## Acceptance Criteria

- [ ] `POST /todos {"title":"..."}` → HTTP 201, body: the created task + `Location: /todos/{id}` header
- [ ] `GET /todos` → HTTP 200, array of all tasks (empty array `[]` if none, not `null`)
- [ ] `GET /todos/1` → HTTP 200, a single task
- [ ] `GET /todos/99` → HTTP 404, `{"error":"task not found"}`
- [ ] `PATCH /todos/1 {"done":true}` → HTTP 200, the updated task
- [ ] `DELETE /todos/1` → HTTP 204, empty body
- [ ] `POST /todos {"title":""}` → HTTP 400, `{"error":"title cannot be empty"}`
- [ ] `GET /todos/abc` → HTTP 400, `{"error":"invalid ID"}`
- [ ] Two concurrent POSTs never create tasks with the same ID
- [ ] Every request is logged: method + path + status + duration
- [ ] `go vet ./...` passes without warnings
- [ ] `go.mod` contains only the `module` and `go` directives

---

## Technical requirements

### Mandatory

| Requirement | Details |
|---|---|
| Router | `http.NewServeMux()` with Go 1.22 patterns: `"GET /todos/{id}"` |
| Path params | `r.PathValue("id")` to extract `{id}` from the path |
| Store | `Store` struct with `map[int]Todo` + `sync.RWMutex` + `next int` |
| Reads | `store.mu.RLock()` / `store.mu.RUnlock()` |
| Writes | `store.mu.Lock()` / `store.mu.Unlock()` (or `defer`) |
| JSON success response | serialized `Todo` or `[]Todo` |
| JSON error response | `{"error":"..."}` |
| HTTP status codes | 200, 201, 204, 400, 404 — strictly per the API table |
| `Content-Type` | `application/json` on every response with a body |
| `Location` header | on `POST /todos` → `w.Header().Set("Location", fmt.Sprintf("/todos/%d", todo.ID))` |
| Empty list | `json.Marshal([]Todo{})` → `[]`, not `null` |
| Middleware | `loggingMiddleware(next http.Handler) http.Handler` |

### Forbidden

- `panic` for error handling
- Third-party packages (`gin`, `echo`, `chi`, `gorilla/mux`)
- Global variables for the store or mutex — only via the `Store` struct
- Direct access to `store.todos` without holding the mutex
- `sync.Mutex` instead of `sync.RWMutex` — a regression for a read-heavy API

---

## Go topics this task exercises

> Not just a list — a checklist of what the implementation **must** use.

- **`sync.RWMutex`** — `RLock/RUnlock` for GET handlers, `Lock/Unlock` for POST/PATCH/DELETE
- **`r.PathValue("id")`** — extracting a path parameter in Go 1.22+, without third-party routers
- **`http.NewServeMux` with methods** — the `"POST /todos"`, `"GET /todos/{id}"` pattern in Go 1.22
- **`w.WriteHeader(http.StatusCreated)`** — setting the status before writing the body
- **`w.Header().Set(...)`** — setting headers before `WriteHeader`
- **`json.NewDecoder(r.Body).Decode`** — streaming decode of the request
- **`json.NewEncoder(w).Encode`** — streaming encode of the response
- **Slice initialization** — `make([]Todo, 0)` instead of `var todos []Todo` to get `[]` instead of `null` in JSON
- **`strconv.Atoi`** — parsing an ID from the path string with error handling
- **`defer mu.Unlock()`** — the deferred-unlock pattern

---

## File structure

```
todo-api/
├── main.go       # NewStore(), handler registration, loggingMiddleware, ListenAndServe
├── store.go      # Todo struct, Store struct, methods: GetAll, GetByID, Create, Update, Delete
├── handler.go    # handlers: listTodos, getTodo, createTodo, updateTodo, deleteTodo
├── middleware.go # loggingMiddleware + responseWriter that captures the status
├── go.mod        # module github.com/Shipovmax/todo-api
└── README.md
```

---

## Architecture hints

```go
// store.go
type Todo struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Done      bool      `json:"done"`
    CreatedAt time.Time `json:"created_at"`
}

type Store struct {
    mu    sync.RWMutex
    todos map[int]Todo
    next  int
}

func NewStore() *Store
func (s *Store) GetAll() []Todo
func (s *Store) GetByID(id int) (Todo, bool)
func (s *Store) Create(title string) Todo
func (s *Store) Update(id int, done bool) (Todo, bool)
func (s *Store) Delete(id int) bool

// handler.go — handlers receive *Store via closure
func createTodo(s *Store) http.HandlerFunc
func listTodos(s *Store) http.HandlerFunc
func getTodo(s *Store) http.HandlerFunc
func updateTodo(s *Store) http.HandlerFunc
func deleteTodo(s *Store) http.HandlerFunc

// main.go
func main() {
    store := NewStore()
    mux := http.NewServeMux()
    mux.HandleFunc("POST /todos", createTodo(store))
    mux.HandleFunc("GET /todos", listTodos(store))
    mux.HandleFunc("GET /todos/{id}", getTodo(store))
    mux.HandleFunc("PATCH /todos/{id}", updateTodo(store))
    mux.HandleFunc("DELETE /todos/{id}", deleteTodo(store))
    http.ListenAndServe(":8080", loggingMiddleware(mux))
}
```

> Handlers via closure (`func createTodo(s *Store) http.HandlerFunc`) — a standard Go pattern for dependency injection without global state. Each handler receives its dependency explicitly.

---

## Definition of Done

1. All acceptance criteria are met
2. Code is pushed to GitHub in the `todo-api` repository
3. README.md in the repository matches the project template
4. You can explain every line of code out loud without looking it up

---

## Next step after review

After review, we move on to **Task #5 — Concurrent Worker**: goroutines, channels, `sync.WaitGroup`, `context.WithCancel` — the fan-out/fan-in task-processing pattern.
