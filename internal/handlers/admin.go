package handlers

import (
	"context"
	"librarymanagementsystem/internal/auth"
	"librarymanagementsystem/internal/db"
	"librarymanagementsystem/internal/models"
	"librarymanagementsystem/templates"
	"net/http"
	"strconv"
)

type AdminHandler struct {
	db             *db.Database
	sessionManager *auth.SessionManager
}

func NewAdminHandler(database *db.Database, sessionManager *auth.SessionManager) *AdminHandler {
	return &AdminHandler{
		db:             database,
		sessionManager: sessionManager,
	}
}

func (h *AdminHandler) Index(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromContext(r.Context())

	// Check admin permission
	hasPerm, err := h.hasPermission(user, "manage_roles")
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}
	if !hasPerm {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	users, err := h.db.GetAllUsersWithRoles()
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	roles, err := h.db.GetAllRoles()
	if err != nil {
		http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
		return
	}

	templates.AdminIndex(users, roles, user).Render(r.Context(), w)
}

func (h *AdminHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromContext(r.Context())

	// Check admin permission
	hasPerm, err := h.hasPermission(user, "manage_roles")
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}
	if !hasPerm {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.FormValue("user_id")
	roleIDStr := r.FormValue("role_id")

	if userIDStr == "" || roleIDStr == "" {
		http.Error(w, "User ID and Role ID are required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		http.Error(w, "Invalid Role ID", http.StatusBadRequest)
		return
	}

	err = h.db.AssignRole(userID, roleID, &user.ID)
	if err != nil {
		http.Error(w, "Failed to assign role", http.StatusInternalServerError)
		return
	}

	// Redirect back to admin page
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromContext(r.Context())

	// Check admin permission
	hasPerm, err := h.hasPermission(user, "manage_roles")
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}
	if !hasPerm {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.FormValue("user_id")
	roleIDStr := r.FormValue("role_id")

	if userIDStr == "" || roleIDStr == "" {
		http.Error(w, "User ID and Role ID are required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		http.Error(w, "Invalid Role ID", http.StatusBadRequest)
		return
	}

	err = h.db.RemoveRole(userID, roleID)
	if err != nil {
		http.Error(w, "Failed to remove role", http.StatusInternalServerError)
		return
	}

	// Redirect back to admin page
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) getUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value("user").(*models.User); ok {
		return user
	}
	return nil
}

func (h *AdminHandler) hasPermission(user *models.User, permissionName string) (bool, error) {
	if user == nil {
		return false, nil
	}
	return h.db.HasPermission(user.ID, permissionName)
}

func (h *AdminHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
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
