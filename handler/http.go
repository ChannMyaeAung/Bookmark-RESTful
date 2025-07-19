package handler

import (
	"Bookmark-RESTful/repository"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Handler provides the database connection to the HTTP handlers.
type Handler struct {
	DB *sql.DB
}

// APIKeyMiddleware validates API keys for protected routes.
func (h *Handler) APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <api-key>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		apiKey := parts[1]
		user, err := repository.GetUserByAPIKey(h.DB, apiKey)
		if err != nil {
			if err == repository.ErrInvalidAPIKey {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Add user to request context
		type contextKey string
		const userContextKey contextKey = "user"
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RegenerateAPIKey handles requests to regenerate a user's API key
// POST /auth/regenerate-key
func (h *Handler) RegenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*repository.User)

	newAPIKey, err := repository.RegenerateAPIKey(h.DB, user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"api_key": newAPIKey,
		"message": "API key regenerated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

// CreateBookmark handles requests to create a new bookmark.
func (h *Handler) CreateBookmark(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*repository.User)

	var bookmark repository.Bookmark
	if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	bm, err := repository.CreateBookmark(h.DB, user.ID, bookmark.Title, bookmark.URL)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bm)
}

// ListBookmarksForCurrentUser handles requests to list current user's bookmarks.
// GET /bookmarks
// Protected Routes (API Key Required)
func (h *Handler) ListBookmarksForCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*repository.User)

	bms, err := repository.FetchBookmarks(h.DB, user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bms)
}

// ListBookmarks handles requests to list a user's bookmarks.
// GET /users/{id}/bookmarks
// Public Routes (No API Key Required)
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
