package utilities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// Rapresents a response
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Success bool        `json:"success"`
}

// Sends HTTP responses
func Reply(w http.ResponseWriter, status int, message string, data interface{}, success bool) {
	response := Response{
		Status:  status,
		Message: message,
		Data:    data,
		Success: success,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// Parses JSON bodies
func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("empty request body")
	}
	return json.NewDecoder(r.Body).Decode(payload)
}

// Creating a new exported *validator.Validate
var Validator = validator.New()

// Registers custom validations
func RegisterCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation("passwordpolicy", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		return len(password) >= 12 && len(password) <= 128 &&
			// At least one uppercase letter
			regexp.MustCompile(`[A-Z]`).MatchString(password) &&
			// At least one lowercase letter
			regexp.MustCompile(`[a-z]`).MatchString(password) &&
			// At least one number
			regexp.MustCompile(`[0-9]`).MatchString(password) &&
			// At least one symbol
			regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]`).MatchString(password)
	})
}

// Validates JSON bodies
func ValidateJSON(w http.ResponseWriter, payload any) map[string]string {
	if err := Validator.Struct(payload); err != nil {
		verr := make(map[string]string)
		var valErr validator.ValidationErrors
		if errors.As(err, &valErr) {
			for _, field := range valErr {
				switch field.Tag() {
				case "required":
					verr[strings.ToLower(field.Field())] = "field required"
				case "email":
					verr[strings.ToLower(field.Field())] = "must be a valid email address"
				case "min":
					verr[strings.ToLower(field.Field())] = fmt.Sprintf("must be at least %s characters", field.Param())
				case "containsuppercase":
					verr[strings.ToLower(field.Field())] = "must contain at least one uppercase letter"
				case "containslowercase":
					verr[strings.ToLower(field.Field())] = "must contain at least one lowercase letter"
				case "containsnumber":
					verr[strings.ToLower(field.Field())] = "must contain at least one number"
				case "containsany":
					verr[strings.ToLower(field.Field())] = fmt.Sprintf("must contain at least one special character (%s)", field.Param())
				case "passwordpolicy":
					verr[strings.ToLower(field.Field())] = fmt.Sprintf("must be at least 12 characters and must contain at least one uppercase letter, lowercase letter and valid symbol")
				default:
					verr[strings.ToLower(field.Field())] = field.Tag()
				}
			}
		}
		return verr
	}
	return nil
}

// Hashes a plain text password
func Hash(plain, pepper string, cost int) (string, error) {
	peppered := plain + pepper
	hash, err := bcrypt.GenerateFromPassword([]byte(peppered), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Compares plain text and hashed password
func ComparePlainAndHashedPassword(hashed, pepper string, plain []byte) bool {
	peppered := string(plain) + pepper
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(peppered))
	return err == nil
}

// Gets the account ID from context
func GetAccountIDFromContext(ctx context.Context) int {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return -1
	}
	// Type switch because JSON defaults to float64
	switch id := claims["id"].(type) {
	case float64:
		return int(id)
	case int:
		return id
	case int64:
		return int(id)
	default:
		return -1
	}
}

// Validetes CORS origins
func ValidateOrigins(origins []string) ([]string, error) {
	var valid []string
	for _, origin := range origins {
		// Trimming extra spaces
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		// Parsing URLs
		u, err := url.Parse(origin)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		// Validating
		if u.Scheme != "http" && u.Scheme != "https" {
			return nil, fmt.Errorf("invalid scheme: %s", origin)
		}
		if u.Host == "" {
			return nil, fmt.Errorf("missing host: %s", origin)
		}
		if u.Path != "" && u.Path != "/" {
			return nil, fmt.Errorf("origin should not contain path: %s", origin)
		}
		valid = append(valid, origin)
	}
	return valid, nil
}
