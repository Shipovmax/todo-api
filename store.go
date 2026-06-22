package main

import (
	"sync"
	"time"
)

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

func NewStore() *Store {
	return &Store{
		todos: make(map[int]Todo),
	}
}

func (s *Store) GetAll() []Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todos := make([]Todo, 0)
	for _, value := range s.todos {
		todos = append(todos, value)
	}
	return todos
}

func (s *Store) GetByID(id int) (Todo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todo, ok := s.todos[id]

	return todo, ok
}

func (s *Store) Create(title string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.next++
	todo := Todo{ID: s.next, Title: title, Done: false, CreatedAt: time.Now()}
	s.todos[todo.ID] = todo
	return todo
}

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
