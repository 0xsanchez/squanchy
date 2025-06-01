package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

// Creating a new enexported *jwtauth.JWTAuth
var tokenAuth *jwtauth.JWTAuth

// Initializes JWT
func (s *Store) InitJWT(secret string) *jwtauth.JWTAuth {
	if secret == "d3f4ult_jwt$$_secret_cha:)nge_me?" {
		fmt.Println("Using the default sample JWT secret please set one for use in production!")
	}
	if len(secret) < 32 {
		fmt.Println("A JWT secret should be at least 32 characters")
	}
	tokenAuth = jwtauth.New("HS256", []byte(secret), nil)
	return tokenAuth
}

// Creates new JWTs
func (s *Store) NewJWT(claims map[string]interface{}) (string, error) {
	expires := time.Now().Add(24 * time.Hour)
	claims["expiration"] = expires.Unix()
	_, token, err := tokenAuth.Encode(claims)
	if err != nil {
		return "", fmt.Errorf("error creating JWT %s", err)
	}
	// Storing the JWT
	_, err = s.db.Exec(`INSERT INTO jwt (token, expires, account) VALUES (?, ?, ?)`, token, expires, claims["id"])
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return "", fmt.Errorf("token collision")
		}
		return "", fmt.Errorf("error storing token: %w", err)
	}
	return token, nil
}

// Invalidates JWTs
func (s *Store) InvalidateSession(token string) error {
	_, err := s.db.Exec("DELETE FROM jwt WHERE token = ?", token)
	return err
}

// Invalidats all JWTs for an account
func (s *Store) InvalidateAllSessions(accountID int) error {
	_, err := s.db.Exec("DELETE FROM jwt WHERE account = ?", accountID)
	return err
}

// Verifies the JWT is active
func (s *Store) IsTokenValid(token string) (exists bool, err error) {
	err = s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM jwt WHERE token = ? AND expires > CURRENT_TIMESTAMP)`, token).Scan(&exists)
	return exists, err
}

// Cleans up expired JWTs
func (s *Store) StartSessionCleanup(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			s.db.Exec("DELETE FROM jwt WHERE expires < CURRENT_TIMESTAMP")
		}
	}()
}
