package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/0xsanchez/squanchy/cmd/squanchy/database"
	"github.com/0xsanchez/squanchy/cmd/squanchy/store"
	"github.com/0xsanchez/squanchy/cmd/squanchy/utilities"
	"github.com/alexflint/go-arg"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joho/godotenv"
)

var args struct {
	Port           int    `arg:"-p,--port,env:PORT" default:"6900" placeholder:"" help:"specifies the port to bind to"`
	Prefix         string `arg:"--prefix,env:PREFIX" default:"/api" placeholder:"" help:"speicifes a prefix"`
	RateLimit      int    `arg:"-r,--soft-rate-limit,env:RATE_LIMIT" default:"60" placeholder:"" help:"specifies the rate limit(1 hour window)"`
	ExtraRateLimit int    `arg:"-e,--hard-rate-limit,env:EXTRA_RATE_LIMIT" default:"10" placeholder:"" help:"specifies the extra rate limit for sensitive endpoints(1 hour window)"`
	Pepper         string `arg:"--pepper,env:PEPPER" default:"" placeholder:"" help:"specifies a password pepper"`
	BcryptCost     int    `arg:"--bcrypt-cost,env:COST" default:"10" placeholder:"" help:"specifies the bcrypt cost"`
	LockTime       int    `arg:"-l,--lock-time,env:LOCK_TIME" default:"6" placeholder:"" help:"specifies the lock time after too many failed attempts in hours"`
	LockAttempts   int    `arg:"-a,--attempts,env:LOCK_ATTEMPTS" default:"9" placeholder:"" help:"specifies the maximum number of failed attempts before being locked out"`
	JwtSecret      string `arg:"-c,--jwt-secret,env:JWT_SECRET" default:"d3f4ult_jwt$$_secret_cha:)nge_me?" placeholder:"" help:"specifies the JWT secret"`
	JwtExpiration  int    `arg:"-x,--jwt-expiration,env:JWT_EXPIRATION" default:"24" placeholder:"" help:"specifies the expiration time of JWTs in hours"`
	DatabasePath   string `arg:"-d,--database,env:DATABASE_PATH" default:"./squanchy.db" placeholder:"" help:"specifies the database"`
	JournalMode    string `arg:"-j,--journal-mode,env:JOURNAL_MODE" default:"DELETE" placeholder:"" help:"specifies the sqlite3 database journal mode"`
	EnableTotp     bool   `arg:"-t,--totp,env:ENABLE_TOTP" default:"false" placeholder:"" help:"specifies whether to enable 2FA using TOTP or not"`
	TotpSkew       uint   `arg:"--totp-skew,env:TOTP_SKEW" default:"0" placeholder:"" help:"specifies the TOTP validation skew"`
	TotpAlgorithm  string `arg:"--totp-algorithm,env:TOTP_ALGORITHM" default:"1" placeholder:"" help:"specifies the TOTP algorithm to use"`
	EnableSmtp     bool   `arg:"-s,--smtp,env:ENABLE_SMTP" default:"false" placeholder:"" help:"specifies whether to enable emails using SMTP"`
	SmtpAddress    string `arg:"--smtp-address,env:SMTP_ADDRESS" placeholder:"" help:"specifies the SMTP server address"`
	SmtpPort       int    `arg:"--smtp-port,env:SMTP_PORT" default:"25"  placeholder:"" help:"specifies the SMTP server port"`
	SmtpUser       string `arg:"--smtp-user,env:SMTP_USER" default:"" placeholder:"" help:"specifies the SMTP server user"`
	SmtpPassword   string `arg:"--smtp-password,env:SMTP_PASSWORD" default:"" placeholder:"" help:"specifies the SMTP server password"`
	SmtpFrom       string `arg:"--smtp-from,env:SMTP_FROM" placeholder:"" help:"specifies the SMTP from address"`
	CorsOrigins    string `arg:"-o,--origins,env:CORS_ORIGINS" default:"" placeholder:"" help:"specifies the CORS origins as a comma separated list"`
	EasterEgg      bool   `arg:"-g,--easter-egg,env:EASTER_EGG" default:"false" placeholder:"" help:"specifies whether to register the easter egg endpoint"`
	Version        bool   `arg:"-v,--version" help:"prints squanchy's version and exit"`
}

//go:embed reference/openapi.yaml
var embedded embed.FS

func init() {
	// Loading enviroment variables 🌹
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading environment variables from")
	}
}

func main() {
	// Parsing CLI flags
	arg.MustParse(&args)
	// Versioning
	if args.Version {
		fmt.Println("squanchy API v0.1")
		os.Exit(0)
	}
	if args.Prefix != "/api" {
		// Validating the prefix
		if args.Prefix[0] != '/' {
			fmt.Println("Error prefix must start with a /")
			os.Exit(1)
		}
	}
	// Connecting to the database
	db, err := database.Connection(args.DatabasePath, args.JournalMode)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()
	// Creating a router
	r := chi.NewRouter()
	// Creating a store
	store := store.NewStore(db)
	// Creating handlers
	handlers := NewHandlers(store)
	// Initializing JWT
	tokenAuth := store.InitJWT(args.JwtSecret)
	// Initializing SMTP
	if args.EnableSmtp {
		if err := store.InitSMTP(args.SmtpAddress, args.SmtpPort, args.SmtpUser, args.SmtpPassword, args.SmtpFrom); err != nil {
		}
	}
	// Registering custom validations
	utilities.RegisterCustomValidators(utilities.Validator)
	// Rate limiting
	r.Use(httprate.LimitByRealIP(args.RateLimit, time.Hour))
	// Validating CORS origins
	origins, err := utilities.ValidateOrigins(strings.Split(args.CorsOrigins, ","))
	if err != nil {
		fmt.Println("Disabling CORS(origins validation failed please provide valid origins)")
		origins = []string{}
	}
	// Setting CORS origins
	custom := ""
	if args.EasterEgg {
		custom = "squanchy"
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", custom},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// Using chi middleware
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	// Using squanchy middleware
	r.Use(AddSecurityHeaders)
	// The easter egg custom method
	chi.RegisterMethod("squanchy")
	// Registering endpoints and serving
	r.Route(args.Prefix, func(r chi.Router) {
		// Unprotected
		r.Group(func(r chi.Router) {
			r.Get("/health", handlers.HealthCheck)
			r.Post("/register", handlers.Register)
			r.Post("/login", handlers.Login)
			// With SMTP enabled
			if args.EnableSmtp {
				r.With(httprate.LimitByIP(args.ExtraRateLimit, time.Hour)).Get("/register/resend", handlers.RegisterResend)
				r.With(httprate.LimitByIP(args.ExtraRateLimit, time.Hour)).Post("/register/confirm", handlers.RegisterConfirm)
				r.With(httprate.LimitByIP(args.ExtraRateLimit, time.Hour)).Post("/recovery", handlers.Recovery)
				r.With(httprate.LimitByIP(args.ExtraRateLimit, time.Hour)).Post("/recovery/confirm", handlers.RecoveryConfirm)
			}
			// With TOTP enabled
			if args.EnableTotp {
				r.With(httprate.LimitByIP(args.ExtraRateLimit, time.Hour)).Post("/recovery/2fa", handlers.RecoveryTOTP)
			}
			// Documentation
			r.Get("/openapi/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
				file, err := embedded.ReadFile("reference/openapi.yaml")
				if err != nil {
					fmt.Println("Error embedding specification", err)
					return
				}
				w.Header().Set("Content-Type", "text/yaml")
				w.Write(file)
			})
			r.Get("/openapi", handlers.OpenAPI)
			// Easter egg
			if args.EasterEgg {
				r.Method("squanchy", "/squanchy", http.HandlerFunc(handlers.squanchy))
			}
		})
		// Protected
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))
			r.Use(Authenticator(store))
			r.Post("/logout", handlers.Logout)
			r.Put("/modify/email", handlers.ChangeEmail)
			r.Put("/modify/password", handlers.ChangePassword)
			r.Delete("/delete", handlers.DeleteAccount)
			// With TOTP enabled
			if args.EnableTotp {
				r.With(httprate.LimitByIP(args.ExtraRateLimit, time.Hour)).Put("/modify/2fa", handlers.Change2FA)
				r.With(httprate.LimitByIP(args.ExtraRateLimit, time.Hour)).Post("/modify/2fa/confirm", handlers.Change2FAConfirm)
			}
		})
	})
	// Starting scheduled cleanup of expired JWTs
	store.StartSessionCleanup(24 * time.Hour)
	// Creating shutdown channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	// Serving
	server := &http.Server{Addr: fmt.Sprintf("localhost:%d", args.Port), Handler: r}
	go func() {
		fmt.Println("Reference at", server.Addr+args.Prefix+"/openapi")
		fmt.Println("Listening to requests on", server.Addr+args.Prefix)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Server error", err)
			os.Exit(1)
		}
	}()
	// Waiting for a shutdown signal
	<-shutdown
	fmt.Println("Shutting down gracefully...")
	// Creating the shutdown context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Attempting a graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Graceful shutdown failed", err)
	}
	fmt.Println("Server stopped")
}
