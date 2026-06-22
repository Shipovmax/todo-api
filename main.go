package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewStore()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /todos", createTodo(store))
	mux.HandleFunc("GET /todos", listTodos(store))
	mux.HandleFunc("GET /todos/{id}", getTodo(store))
	mux.HandleFunc("PATCH /todos/{id}", updateTodo(store))
	mux.HandleFunc("DELETE /todos/{id}", deleteTodo(store))
	if err := http.ListenAndServe(":8080", loggingMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}
