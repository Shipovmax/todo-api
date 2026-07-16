package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// writeJSON serializes v as JSON, sets the Content-Type header and the
// given HTTP status, and writes the result to w. Encoding failures are
// logged since the status line has already been sent and cannot be
// retried or surfaced to the client at this point.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: failed to encode response body: %v", err)
	}
}

// listTodos returns a handler that responds with all tasks currently
// held by the store.
func listTodos(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.GetAll())
	}
}

// getTodo returns a handler that responds with a single task identified
// by the "id" path parameter.
func getTodo(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ID"})
			return
		}
		todo, ok := s.GetByID(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		writeJSON(w, http.StatusOK, todo)
	}
}

// createTodo returns a handler that creates a new task from the JSON
// request body ({"title": "..."}) and responds with the created task.
func createTodo(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if req.Title == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title cannot be empty"})
			return
		}
		todo := s.Create(req.Title)
		w.Header().Set("Location", fmt.Sprintf("/todos/%d", todo.ID))
		writeJSON(w, http.StatusCreated, todo)
	}
}

// updateTodo returns a handler that updates the "done" field of the task
// identified by the "id" path parameter from the JSON request body
// ({"done": true|false}) and responds with the updated task.
func updateTodo(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ID"})
			return
		}
		var req struct {
			Done bool `json:"done"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		todo, ok := s.Update(id, req.Done)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		writeJSON(w, http.StatusOK, todo)
	}
}

// deleteTodo returns a handler that deletes the task identified by the
// "id" path parameter and responds with 204 No Content on success.
func deleteTodo(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ID"})
			return
		}
		ok := s.Delete(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
