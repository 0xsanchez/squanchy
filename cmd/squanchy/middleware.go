package main

import (
	"net/http"
	"strings"

	"github.com/0xsanchez/squanchy/cmd/squanchy/store"
	"github.com/0xsanchez/squanchy/cmd/squanchy/utilities"
)

// Adds security headers
func AddSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		// w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; form-action 'self'; object-src 'none'")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("X-XSS-Protection", "1; mode=block") // Obsolete in modern browsers
		w.Header().Del("Server")
		next.ServeHTTP(w, r)
	})
}

// Protects endpoints from request with invalidated JWTs
func Authenticator(store *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
			valid, err := store.IsTokenValid(tokenString)
			if err != nil || !valid {
				utilities.Reply(w, http.StatusUnauthorized, "invalid token", nil, false)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
