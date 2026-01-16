package main

import (
	"fmt"
	"net/http"
)

type APIHandler struct {
	app *App
}

func NewAPIHandler(app *App) *APIHandler {
	return &APIHandler{app: app}
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/greet":
		h.handleGreet(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *APIHandler) handleGreet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		name = "World"
	}

	greeting := h.app.Greet(name)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<span class="text-success">%s</span>`, greeting)
}
