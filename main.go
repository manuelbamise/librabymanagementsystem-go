package main

import (
	"librarymanagementsystem/internal/auth"
	"librarymanagementsystem/internal/db"
	"librarymanagementsystem/internal/handlers"
	"librarymanagementsystem/templates"
	"log"
	"net/http"
	"time"
)

func main() {
	// Initialize database
	database, err := db.NewDatabase("library.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Initialize session manager
	sessionManager := auth.NewSessionManager()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(database, sessionManager)
	libraryHandler := handlers.NewLibraryHandler(database, sessionManager)
	adminHandler := handlers.NewAdminHandler(database, sessionManager)

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Home page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// Check if user is logged in
			sessionToken, err := auth.GetSessionToken(r)
			if err == nil {
				if _, err := sessionManager.ValidateSession(sessionToken); err == nil {
					// User is logged in, redirect to library
					http.Redirect(w, r, "/library", http.StatusSeeOther)
					return
				}
			}
			// Show landing page
			templates.Landing().Render(r.Context(), w)
		} else {
			http.NotFound(w, r)
		}
	})

	// Authentication routes
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/logout", authHandler.Logout)

	// Library routes (protected)
	mux.HandleFunc("/library", libraryHandler.AuthMiddleware(libraryHandler.Index))
	mux.HandleFunc("/library/search", libraryHandler.AuthMiddleware(libraryHandler.Search))
	mux.HandleFunc("/library/view/", libraryHandler.AuthMiddleware(libraryHandler.ViewPDF))
	mux.HandleFunc("/upload", libraryHandler.AuthMiddleware(libraryHandler.UploadForm))
	mux.HandleFunc("/library/upload", libraryHandler.AuthMiddleware(libraryHandler.UploadPDF))
	mux.HandleFunc("/library/delete", libraryHandler.AuthMiddleware(libraryHandler.DeletePDF))

	// Admin routes (protected)
	mux.HandleFunc("/admin", adminHandler.AuthMiddleware(adminHandler.Index))
	mux.HandleFunc("/admin/assign-role", adminHandler.AuthMiddleware(adminHandler.AssignRole))
	mux.HandleFunc("/admin/remove-role", adminHandler.AuthMiddleware(adminHandler.RemoveRole))

	// Start session cleanup goroutine
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			sessionManager.CleanupExpiredSessions()
		}
	}()

	// Start server
	log.Println("Server starting on :8009")
	log.Fatal(http.ListenAndServe(":8009", mux))
}
