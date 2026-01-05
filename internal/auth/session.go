package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"
)

var (
	ErrNoAuthHeader    = errors.New("no authorization header")
	ErrInvalidAuthType = errors.New("invalid authorization type")
	ErrInvalidToken    = errors.New("invalid token")
)

type Session struct {
	UserID    int
	Username  string
	ExpiresAt time.Time
}

type SessionManager struct {
	sessions map[string]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) CreateSession(userID int, username string) (string, error) {
	token := generateToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	sm.sessions[token] = &Session{
		UserID:    userID,
		Username:  username,
		ExpiresAt: expiresAt,
	}

	return token, nil
}

func (sm *SessionManager) ValidateSession(token string) (*Session, error) {
	session, exists := sm.sessions[token]
	if !exists {
		return nil, ErrInvalidToken
	}

	if time.Now().After(session.ExpiresAt) {
		delete(sm.sessions, token)
		return nil, ErrInvalidToken
	}

	return session, nil
}

func (sm *SessionManager) DestroySession(token string) {
	delete(sm.sessions, token)
}

func (sm *SessionManager) CleanupExpiredSessions() {
	now := time.Now()
	for token, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, token)
		}
	}
}

func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		Secure:   false, // Set to true in production with HTTPS
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func GetSessionToken(r *http.Request) (string, error) {
	// First try to get from cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		return cookie.Value, nil
	}

	// Fallback to Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeader
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidAuthType
	}

	return parts[1], nil
}
