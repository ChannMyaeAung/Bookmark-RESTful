package main

import (
	"Bookmark-RESTful/db"
	"Bookmark-RESTful/handler"
	"Bookmark-RESTful/repository"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system variables.")
	}

	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer database.Close()
	fmt.Println("Successfully connected to the database!")

	// Generate API keys for existing users who don't have one
	if err := repository.UpdateExistingUsersWithAPIKey(database); err != nil {
		log.Printf("Warning: Could not update existing users with API keys: %v\n", err)
	} else {
		log.Println("Successfully updated existing users with API keys.")
	}

	// Initialize the handler with the database connection
	h := &handler.Handler{DB: database}

	// Create a new chi router
	r := chi.NewRouter()

	// Add a logger middleware to see request details in the console
	r.Use(middleware.Logger)

	// Public routes (no API key required)
	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.CreateUser)

		r.Route("/{userID}", func(r chi.Router) {
			r.Get("/bookmarks", h.ListBookmarks)
		})
	})

	// Protected routes (API key required)
	r.Route("/", func(r chi.Router) {
		// apply the API key middleware to all routes in this group
		r.Use(h.APIKeyMiddleware)

		r.Route("/bookmarks", func(r chi.Router) {
			r.Post("/", h.CreateBookmark)
			r.Get("/", h.ListBookmarksForCurrentUser)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/regenerate-key", h.RegenerateAPIKey)
		})
	})

	// Start the server
	port := ":8080"
	log.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
