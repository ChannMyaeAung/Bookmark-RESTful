package handler

import (
	"Bookmark-RESTful/repository"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Handler provides the database connection to the HTTP handlers.
type Handler struct {
	DB *sql.DB
}

// CreateUser handles requests to create a new user.
// POST /users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u repository.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := repository.CreateUser(h.DB, u.Name, u.Email)
	if err != nil {
		if err == repository.ErrEmailTaken {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(user)
}

// ListBookmarks handles requests to list a user's bookmarks.
// GET /users/{id}/bookmarks
func (h *Handler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	// chi.URLParam to get the userID from the path
	userIDStr := chi.URLParam(r, "userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	bms, err := repository.FetchBookmarks(h.DB, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bms)
}
