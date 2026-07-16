package main

import (
	"sync"
	"time"
)

// Todo represents a single task.
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

// Store is an in-memory, thread-safe collection of tasks. The zero value
// is not usable; construct one with NewStore. All access to the
// underlying map goes through Store's methods, which guard it with mu:
// RLock/RUnlock for reads, Lock/Unlock for writes.
type Store struct {
	mu    sync.RWMutex
	todos map[int]Todo
	next  int
}

// NewStore creates an empty, ready-to-use Store.
func NewStore() *Store {
	return &Store{
		todos: make(map[int]Todo),
	}
}

// GetAll returns a snapshot of all tasks currently in the store. The
// returned slice is never nil, so it serializes to "[]" rather than
// "null" when there are no tasks.
func (s *Store) GetAll() []Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todos := make([]Todo, 0)
	for _, value := range s.todos {
		todos = append(todos, value)
	}
	return todos
}

// GetByID returns the task with the given ID and true, or a zero Todo
// and false if no such task exists.
func (s *Store) GetByID(id int) (Todo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todo, ok := s.todos[id]

	return todo, ok
}

// Create allocates a new task with the given title, assigns it the next
// available ID, and stores it.
func (s *Store) Create(title string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.next++
	todo := Todo{ID: s.next, Title: title, Done: false, CreatedAt: time.Now()}
	s.todos[todo.ID] = todo
	return todo
}

// Update sets the Done field of the task with the given ID and returns
// the updated task and true. It returns a zero Todo and false if no task
// with that ID exists.
func (s *Store) Update(id int, done bool) (Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, ok := s.todos[id]
	if !ok {
		return Todo{}, false
	}
	todo.Done = done
	s.todos[id] = todo

	return todo, true
}

// Delete removes the task with the given ID and reports whether a task
// with that ID existed.
func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.todos[id]
	if !ok {
		return false
	}

	delete(s.todos, id)
	return true
}
