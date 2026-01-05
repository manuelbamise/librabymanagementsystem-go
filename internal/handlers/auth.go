package handlers

import (
	"context"
	"fmt"
	"librarymanagementsystem/internal/auth"
	"librarymanagementsystem/internal/db"
	"librarymanagementsystem/templates"
	"net/http"
	"strings"
)

type AuthHandler struct {
	db             *db.Database
	sessionManager *auth.SessionManager
}

func NewAuthHandler(database *db.Database, sessionManager *auth.SessionManager) *AuthHandler {
	return &AuthHandler{
		db:             database,
		sessionManager: sessionManager,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.Login().Render(r.Context(), w)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		h.renderAuthError(w, "Username and password are required")
		return
	}

	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		h.renderAuthError(w, "Invalid username or password")
		return
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		h.renderAuthError(w, "Invalid username or password")
		return
	}

	sessionToken, err := h.sessionManager.CreateSession(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	auth.SetSessionCookie(w, sessionToken)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/library")
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/library", http.StatusSeeOther)
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.Register().Render(r.Context(), w)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm-password")

	if username == "" || email == "" || password == "" {
		h.renderAuthError(w, "All fields are required")
		return
	}

	if password != confirmPassword {
		h.renderAuthError(w, "Passwords do not match")
		return
	}

	if len(password) < 6 {
		h.renderAuthError(w, "Password must be at least 6 characters")
		return
	}

	// Check if user already exists
	_, err := h.db.GetUserByUsername(username)
	if err == nil {
		h.renderAuthError(w, "Username already exists")
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create user
	err = h.db.CreateUser(username, email, passwordHash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.renderAuthError(w, "Username or email already exists")
		} else {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}
		return
	}

	// Auto-login after registration
	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	sessionToken, err := h.sessionManager.CreateSession(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	auth.SetSessionCookie(w, sessionToken)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/library")
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/library", http.StatusSeeOther)
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := auth.GetSessionToken(r)
	if err == nil {
		h.sessionManager.DestroySession(sessionToken)
	}

	auth.ClearSessionCookie(w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) renderAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)

	errorHTML := fmt.Sprintf(`<div class="error-message">%s</div>`, message)
	w.Write([]byte(errorHTML))
}

func (h *AuthHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionToken, err := auth.GetSessionToken(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, err := h.sessionManager.ValidateSession(sessionToken)
		if err != nil {
			auth.ClearSessionCookie(w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user with roles to context
		user, err := h.db.GetUserWithRoles(session.UserID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
