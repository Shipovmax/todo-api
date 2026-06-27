# Task #4 — TODO HTTP API

## Цель

Написать REST API для управления задачами на чистом `net/http` с in-memory хранилищем. Главная учебная цель — освоить полный CRUD через HTTP, роутинг с параметрами пути в Go 1.22+, и потокобезопасный доступ к данным через `sync.RWMutex`. Конкурентный доступ к shared state — одна из главных тем на Go-собеседованиях в Ozon/WB/Сбер.

---

## Acceptance Criteria

- [ ] `POST /todos {"title":"..."}` → HTTP 201, тело: созданная задача + `Location: /todos/{id}` header
- [ ] `GET /todos` → HTTP 200, массив всех задач (пустой массив `[]` если нет задач, не `null`)
- [ ] `GET /todos/1` → HTTP 200, одна задача
- [ ] `GET /todos/99` → HTTP 404, `{"error":"задача не найдена"}`
- [ ] `PATCH /todos/1 {"done":true}` → HTTP 200, обновлённая задача
- [ ] `DELETE /todos/1` → HTTP 204, пустое тело
- [ ] `POST /todos {"title":""}` → HTTP 400, `{"error":"title не может быть пустым"}`
- [ ] `GET /todos/abc` → HTTP 400, `{"error":"некорректный ID"}`
- [ ] Два параллельных POST не создают задачи с одинаковым ID
- [ ] Каждый запрос логируется: метод + путь + статус + duration
- [ ] `go vet ./...` проходит без предупреждений
- [ ] `go.mod` содержит только `module` и `go` директивы

---

## Технические требования

### Обязательно

| Требование | Детали |
|---|---|
| Роутер | `http.NewServeMux()` с паттернами Go 1.22: `"GET /todos/{id}"` |
| Path params | `r.PathValue("id")` для извлечения `{id}` из пути |
| Хранилище | `Store` struct с `map[int]Todo` + `sync.RWMutex` + `next int` |
| Чтение | `store.mu.RLock()` / `store.mu.RUnlock()` |
| Запись | `store.mu.Lock()` / `store.mu.Unlock()` (или `defer`) |
| JSON ответ успех | сериализованный `Todo` или `[]Todo` |
| JSON ответ ошибка | `{"error":"..."}` |
| HTTP статусы | 200, 201, 204, 400, 404 — строго по таблице API |
| `Content-Type` | `application/json` на все ответы с телом |
| `Location` header | при `POST /todos` → `w.Header().Set("Location", fmt.Sprintf("/todos/%d", todo.ID))` |
| Пустой список | `json.Marshal([]Todo{})` → `[]`, не `null` |
| Middleware | `loggingMiddleware(next http.Handler) http.Handler` |

### Запрещено

- `panic` для обработки ошибок
- Сторонние пакеты (`gin`, `echo`, `chi`, `gorilla/mux`)
- Глобальные переменные для хранилища или mutex — только через `Store` struct
- Прямой доступ к `store.todos` без блокировки mutex
- `sync.Mutex` вместо `sync.RWMutex` — для read-heavy API это регресс

---

## Темы Go, которые ты прокачиваешь

> Это не просто список — это checklist того, что **обязан использовать** в реализации.

- **`sync.RWMutex`** — `RLock/RUnlock` для GET-хендлеров, `Lock/Unlock` для POST/PATCH/DELETE
- **`r.PathValue("id")`** — извлечение параметра пути в Go 1.22+, без сторонних роутеров
- **`http.NewServeMux` с методами** — паттерн `"POST /todos"`, `"GET /todos/{id}"` в Go 1.22
- **`w.WriteHeader(http.StatusCreated)`** — установка статуса до записи тела
- **`w.Header().Set(...)`** — установка заголовков до `WriteHeader`
- **`json.NewDecoder(r.Body).Decode`** — стриминговый decode запроса
- **`json.NewEncoder(w).Encode`** — стриминговый encode ответа
- **Инициализация slice** — `make([]Todo, 0)` вместо `var todos []Todo` чтобы получить `[]` вместо `null` в JSON
- **`strconv.Atoi`** — парсинг ID из строки пути с обработкой ошибки
- **`defer mu.Unlock()`** — паттерн отложенной разблокировки

---

## Структура файлов

```
todo-api/
├── main.go       # NewStore(), регистрация хендлеров, loggingMiddleware, ListenAndServe
├── store.go      # Todo struct, Store struct, методы: GetAll, GetByID, Create, Update, Delete
├── handler.go    # хендлеры: listTodos, getTodo, createTodo, updateTodo, deleteTodo
├── middleware.go # loggingMiddleware + responseWriter с перехватом статуса
├── go.mod        # module github.com/Shipovmax/todo-api
└── README.md
```

---

## Подсказки по архитектуре

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

// handler.go — хендлеры принимают *Store через замыкание
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

> Хендлеры через замыкание (`func createTodo(s *Store) http.HandlerFunc`) — стандартный Go-паттерн для dependency injection без глобального состояния. Каждый хендлер получает зависимость явно.

---

## Definition of Done

1. Все Acceptance Criteria выполнены
2. Код запушен на GitHub в репозиторий `todo-api`
3. README.md в репозитории соответствует шаблону проекта
4. Ты можешь объяснить каждую строку кода вслух без подглядывания

---

## Следующий шаг после сдачи

После ревью переходим к **Task #5 — Конкурентный воркер**: goroutines, channels, `sync.WaitGroup`, `context.WithCancel` — fan-out/fan-in паттерн обработки задач.
