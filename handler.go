package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func listTodos(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.GetAll())
	}
}

func getTodo(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный ID"})
			return
		}
		todo, ok := s.GetByID(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "задача не найдена"})
			return
		}
		writeJSON(w, http.StatusOK, todo)
	}
}

func createTodo(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный JSON"})
			return
		}
		if req.Title == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title не может быть пустым"})
			return
		}
		todo := s.Create(req.Title)
		w.Header().Set("Location", fmt.Sprintf("/todos/%d", todo.ID))
		writeJSON(w, http.StatusCreated, todo)
	}
}

func updateTodo(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный ID"})
			return
		}
		var req struct {
			Done bool `json:"done"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный JSON"})
			return
		}
		todo, ok := s.Update(id, req.Done)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "задача не найдена"})
			return
		}
		writeJSON(w, http.StatusOK, todo)
	}
}

func deleteTodo(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный ID"})
			return
		}
		ok := s.Delete(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "задача не найдена"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
