package middleware

import (
	"context"
	"net/http"

	"github.com/siddhantagarwal/expense-manager/internal/auth"
)

type contextKey string

const userIDKey contextKey = "user_id"

func NewContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func FromContext(ctx context.Context) (string, bool) {
	s, ok := ctx.Value(userIDKey).(string)
	return s, ok
}

func AuthMiddleware(a *auth.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := auth.GetSessionToken(r)
			if token == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			session, ok := a.GetSession(token)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Store username in request context for handlers to use
			next.ServeHTTP(w, r.WithContext(NewContext(r.Context(), session.Username)))
		})
	}
}
