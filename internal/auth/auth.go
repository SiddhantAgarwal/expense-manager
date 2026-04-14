package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	Username  string
	ExpiresAt time.Time
}

type Auth struct {
	mu       sync.Mutex
	sessions map[string]Session
}

func New() *Auth {
	return &Auth{
		sessions: make(map[string]Session),
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (a *Auth) CreateSession(username string) string {
	a.mu.Lock()
	defer a.mu.Unlock()

	token := generateToken()
	a.sessions[token] = Session{
		Username:  username,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	return token
}

func (a *Auth) GetSession(token string) (Session, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	session, ok := a.sessions[token]
	if !ok || time.Now().After(session.ExpiresAt) {
		delete(a.sessions, token)
		return Session{}, false
	}

	return session, true
}

func (a *Auth) DeleteSession(token string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.sessions, token)
}

func (a *Auth) CleanExpired() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	for token, session := range a.sessions {
		if now.After(session.ExpiresAt) {
			delete(a.sessions, token)
		}
	}
}

func GetSessionToken(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}

	return cookie.Value
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}
