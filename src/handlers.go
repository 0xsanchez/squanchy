package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/0xsanchez/squanchy/src/store"
	"github.com/0xsanchez/squanchy/src/types"
	"github.com/0xsanchez/squanchy/src/utilities"
	"github.com/MarceloPetrucio/go-scalar-api-reference"
)

// ------------------------------------------------------------------------------------------------------------
// All the handlers.                                                                       			          |
// I reccommend mounting the API under a versioned prefix using --prefix or the PREFIX environment variable!  |
// In alternative you can do the same when using a reverse proxy such as Nginx and Caddy.			  		  |
// ------------------------------------------------------------------------------------------------------------

type Handlers struct {
	store *store.Store
}

func NewHandlers(store *store.Store) *Handlers {
	return &Handlers{
		store: store,
	}
}

// POST /register
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.RegisterPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusBadRequest, "invalid payload", verr, false)
		return
	}
	// Hashing the password
	hashedPassword, err := utilities.Hash(payload.Password, args.Pepper, args.BcryptCost)
	if err != nil {
		fmt.Println("Error hashing password", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Is an account with that email already existing?
	if _, err := h.store.GetAccountByEmail(payload.Email); err == nil {
		utilities.Reply(w, http.StatusConflict, "email already exists", nil, false)
		return
	}
	if err := h.store.NewAccount(payload.Email, hashedPassword); err != nil {
		fmt.Println("Error creating account", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	if args.EnableSmtp {
		// Getting the account by email
		account, err := h.store.GetAccountByEmail(payload.Email)
		if err != nil {
			utilities.Reply(w, http.StatusNotFound, "account not found", nil, false)
			return
		}
		// Generating a new email verification code
		code, err := h.store.GenerateEmailVerificationCode()
		if err != nil {
			fmt.Println("Error generating email verification code", err)
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
		// Storing the email verification code
		if err := h.store.NewEmailVerificationCode(account.ID, code); err != nil {
			fmt.Println("Error inserting email verification code in the database", err)
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
		// Sending a verification email
		if err := h.store.SendEmail(payload.Email, "Verification", fmt.Sprintf("The verification code is %s", code)); err != nil {
			fmt.Println("Error sending email", err)
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
		utilities.Reply(w, http.StatusOK, "verification email sent", nil, true)
	} else {
		utilities.Reply(w, http.StatusCreated, "account created", nil, true)
	}
}

// GET /register/resend, SMTP server only
func (h *Handlers) RegisterResend(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.RegisterResendPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusBadRequest, "invalid payload", verr, false)
		return
	}
	// Getting the account by email
	account, err := h.store.GetAccountByEmail(payload.Email)
	if err != nil {
		utilities.Reply(w, http.StatusNotFound, "account not found", nil, false)
		return
	}
	// Is the account already verified?
	if account.Verified {
		utilities.Reply(w, http.StatusConflict, "already verified", nil, false)
		return
	}
	// Generating a new email verification code
	code, err := h.store.GenerateEmailVerificationCode()
	if err != nil {
		fmt.Println("Error generating email verification code", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Storing the email verification code
	if err := h.store.NewEmailVerificationCode(account.ID, code); err != nil {
		fmt.Println("Error inserting email verification code in the database", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Sending a verification email
	if err := h.store.SendEmail(payload.Email, "Verification", fmt.Sprintf("The verification code is %s", code)); err != nil {
		fmt.Println("Error sending email", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	utilities.Reply(w, http.StatusOK, "verification email resent", nil, true)
}

// POST /register/confirm, SMTP server only
func (h *Handlers) RegisterConfirm(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.RegisterConfirmPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusBadRequest, "invalid payload", verr, false)
		return
	}
	// Getting the account by email
	account, err := h.store.GetAccountByEmail(payload.Email)
	if err != nil {
		utilities.Reply(w, http.StatusNotFound, "account not found", nil, false)
		return
	}
	// Is the account already verified?
	if account.Verified {
		utilities.Reply(w, http.StatusConflict, "already verified", nil, false)
		return
	}
	// Is the email verification code valid?
	valid, err := h.store.IsEmailVerificationCodeValid(account.ID, payload.Code)
	if err != nil {
		fmt.Println("Error verifying email verification code", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	if !valid {
		utilities.Reply(w, http.StatusUnauthorized, "invalid email verification code", nil, false)
		return
	}
	// Verifing the account
	if err := h.store.VerifyAccount(account.ID); err != nil {
		fmt.Print("Error verifying account", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Cleaning up verification codes
	if err := h.store.DeleteEmailVerificationCodes(account.ID); err != nil {
		fmt.Println("Error cleaning up verification codes", err)
	}
	// Sending a notification email
	if err := h.store.SendEmail(payload.Email, "Welcome", "Your email has been verified correctly"); err != nil {
		fmt.Println("Error sending email", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	utilities.Reply(w, http.StatusOK, "verified", nil, true)
}

// POST /login
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.LoginPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusBadRequest, "invalid payload", verr, false)
		return
	}
	// Getting the account by email
	account, err := h.store.GetAccountByEmail(payload.Email)
	if err != nil {
		utilities.Reply(w, http.StatusUnauthorized, "email not found", nil, false)
		return
	}
	// Is SMTP enabled?
	if args.EnableSmtp {
		// Is the account verified?
		if !account.Verified {
			utilities.Reply(w, http.StatusUnauthorized, "account unverified", nil, false)
			return
		}
	}
	// Is TOTP enabled?
	if args.EnableTotp {
		// Is account 2FA enabled?
		if account.TOTP {
			// Verifing the TOTP code
			valid, err := h.store.ValidateTOTP(account.ID, payload.TOTP, args.TotpAlgorithm, args.TotpSkew)
			if err != nil {
				fmt.Println("Error validating TOTP", err)
				utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
				return
			}
			if !valid {
				utilities.Reply(w, http.StatusUnauthorized, "invalid TOTP code", nil, false)
				return
			}
		}
	}
	// Is the account locked?
	if account.LockedUntil.After(time.Now()) {
		utilities.Reply(w, http.StatusTooManyRequests, "account locked", nil, false)
		return
	}
	// Comparing the plaintext and hashed passwords
	if !utilities.ComparePlainAndHashedPassword(account.Password, args.Pepper, []byte(payload.Password)) {
		// Tracking login attempts
		attempts := account.FailedAttempts + 1
		if err := h.store.TrackLoginAttempts(account.ID, attempts); err != nil {
			fmt.Println("Error trackign login attempts", err)
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
		// Locking the account
		if attempts >= args.LockAttempts {
			lockDuration := time.Duration(args.LockTime) * time.Hour
			lockTime := time.Now().Add(lockDuration)
			if err := h.store.LockAccount(account.ID, lockTime); err != nil {
				fmt.Println("Error locking account", err)
				utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
				return
			}
			fmt.Println("Account locked with ID", account.ID, "and email", account.Email, "for", args.LockTime, "hours")
			if err := h.store.ResetLoginAttempts(account.ID); err != nil {
				fmt.Println("Error resetting login attempts", err)
				utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
				return
			}
			utilities.Reply(w, http.StatusForbidden, "account locked due to too many failed attepts", nil, false)
			return
		}
		utilities.Reply(w, http.StatusUnauthorized, "incorrect email or password", nil, false)
		return
	}
	// Tracking last login
	if err := h.store.ChangeLastLogin(account.ID); err != nil {
		fmt.Println("Error changing last login", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Creating JWT claims
	claims := map[string]interface{}{
		"id":         account.ID,
		"issued":     time.Now().Unix(),
		"expiration": time.Now().Add(time.Duration(args.JwtExpiration) * time.Hour).Unix(),
	}
	// Creating a new JWT
	token, err := h.store.NewJWT(claims)
	if err != nil {
		utilities.Reply(w, http.StatusUnauthorized, "invalid claims", nil, false)
		return
	}
	// Storing the JWT
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(time.Duration(args.JwtExpiration) * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	utilities.Reply(w, http.StatusOK, "token", map[string]interface{}{"token": token}, true)
}

// POST /logout
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// Getting the token
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	// Invalidating the token
	if err := h.store.InvalidateSession(token); err != nil {
		fmt.Println("Error invalidating session", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Clearing the client authentication cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	utilities.Reply(w, http.StatusOK, "logged out", nil, true)
}

// POST /recovery, SMTP server only
func (h *Handlers) Recovery(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.RecoveryPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusBadRequest, "invalid payload", verr, false)
		return
	}
	// Is an account with that email already existing?
	account, err := h.store.GetAccountByEmail(payload.Email)
	if err != nil {
		if err.Error() == "account not found" {
			utilities.Reply(w, http.StatusUnauthorized, "email not found", nil, false)
			return
		}
		fmt.Println("Error getting account by email", err)
		utilities.Reply(w, http.StatusConflict, "email already exists", nil, false)
		return
	}
	// Generating a new email recovery code
	code, err := h.store.GenerateEmailVerificationCode()
	if err != nil {
		fmt.Println("Error generating email recovery code", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Storing the email recovery code
	if err := h.store.NewEmailVerificationCode(account.ID, code); err != nil {
		fmt.Println("Error inserting email verification code in the database", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Sending a recovery email
	if err := h.store.SendEmail(payload.Email, "Recovery", fmt.Sprintf("The recovery code is %s", code)); err != nil {
		fmt.Println("Error sending email", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	utilities.Reply(w, http.StatusOK, "recovery code sent", nil, true)
}

// POST /recovery/confirm, SMTP server only
func (h *Handlers) RecoveryConfirm(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.RecoveryConfirmPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusBadRequest, "invalid payload", verr, false)
		return
	}
	// Getting the account by email
	account, err := h.store.GetAccountByEmail(payload.Email)
	if err != nil {
		if err.Error() == "account not found" {
			utilities.Reply(w, http.StatusUnauthorized, "email not found", nil, false)
			return
		}
		fmt.Println("Error getting account by email", err)
		utilities.Reply(w, http.StatusConflict, "email already exists", nil, false)
		return
	}
	// Is the email verification code valid?
	valid, err := h.store.IsEmailVerificationCodeValid(account.ID, payload.Code)
	if err != nil {
		fmt.Println("Error verifying email verification code", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	if !valid {
		utilities.Reply(w, http.StatusUnauthorized, "invalid email verification code", nil, false)
		return
	}
	// Hashing the new password
	newHashedPassword, err := utilities.Hash(payload.New, args.Pepper, args.BcryptCost)
	if err != nil {
		fmt.Println("Error hashing password", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Changing the password
	err = h.store.ChangeAccountPassword(account.ID, newHashedPassword)
	if err != nil {
		fmt.Println("Error changing account password", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	utilities.Reply(w, http.StatusOK, "account recovered", nil, true)
}

// POST /recovery/totp, 2FA TOTP only
func (h *Handlers) RecoveryTOTP(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.RecoveryTOTPPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusBadRequest, "invalid payload", verr, false)
		return
	}
	// Getting the account by email
	account, err := h.store.GetAccountByEmail(payload.Email)
	if err != nil {
		if err.Error() == "account not found" {
			utilities.Reply(w, http.StatusUnauthorized, "email not found", nil, false)
			return
		}
		fmt.Println("Error getting account by email", err)
		utilities.Reply(w, http.StatusConflict, "email already exists", nil, false)
		return
	}
	// Is the account 2FA disabled?
	if !account.TOTP {
		fmt.Println("Warning someone is likely penetration testing")
		utilities.Reply(w, http.StatusUnauthorized, "2FA already disabled", nil, false)
		return
	}
	// Is the TOTP backup code valid?
	valid, err := h.store.IsTOTPBackupCodeValid(account.ID, payload.Code)
	if err != nil {
		fmt.Println("Error verifing TOTP backup code", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	if !valid {
		utilities.Reply(w, http.StatusUnauthorized, "invalid backup code", nil, false)
		return
	}
	// Disabling account 2FA
	if err := h.store.ToggleAccount2FA(account.ID, account.TOTP); err != nil {
		fmt.Println("Error changing account 2FA status", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Sending a notification email
	if args.EnableSmtp {
		if err := h.store.SendEmail(account.Email, "Notification", "2FA disabled"); err != nil {
			fmt.Println("Error sending email", err)
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
	}
	utilities.Reply(w, http.StatusOK, "2fa disabled", nil, true)
}

// PUT /modify/email
func (h *Handlers) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.ChangeEmailPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusBadRequest, "invalid payload", verr, false)
		return
	}
	// Getting the account ID from context
	id := utilities.GetAccountIDFromContext(r.Context())
	if id == -1 {
		fmt.Println("Error getting account ID from context")
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Getting the account by ID
	account, err := h.store.GetAccountByID(id)
	if err != nil {
		fmt.Println("Error getting account by ID", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Is an account with that email already existing?
	if _, err := h.store.GetAccountByEmail(payload.New); err == nil {
		fmt.Println("Error getting account by email", err)
		utilities.Reply(w, http.StatusConflict, "email already exists", nil, false)
		return
	}
	// Changing the email
	if err := h.store.ChangeAccountEmail(account.ID, payload.New); err != nil {
		fmt.Println("Error changing account email", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Sending a notification email
	if args.EnableSmtp {
		if err := h.store.SendEmail(account.Email, "Notification", fmt.Sprintf("Your email address has been modified to %s", payload.New)); err != nil {
			fmt.Println("Error sending email", err)
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
		// Sending another notification email
		if err := h.store.SendEmail(payload.New, "Notification", "This is your new email address"); err != nil {
			fmt.Println("Error sending email", err)
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
	}
	utilities.Reply(w, http.StatusOK, "email updated", nil, true)
}

// PUT /modify/password
func (h *Handlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.ChangePasswordPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusUnprocessableEntity, "invalid payload", verr, false)
		return
	}
	// Getting the account ID from context
	id := utilities.GetAccountIDFromContext(r.Context())
	if id == -1 {
		fmt.Println("Error getting account ID from context")
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Getting the account by ID
	account, err := h.store.GetAccountByID(id)
	if err != nil {
		fmt.Println("Error getting account by ID")
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Comparing the plaintext and hashed passwords
	if !utilities.ComparePlainAndHashedPassword(account.Password, args.Pepper, []byte(payload.Password)) {
		utilities.Reply(w, http.StatusUnauthorized, "wrong password", nil, false)
		return
	}
	// Is the new password different?
	if payload.Password == payload.New {
		utilities.Reply(w, http.StatusBadRequest, "new password must be different", nil, false)
		return
	}
	// Hashing the new password
	newHashedPassword, err := utilities.Hash(payload.New, args.Pepper, args.BcryptCost)
	if err != nil {
		fmt.Println("Error hashing password", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Changing the password
	err = h.store.ChangeAccountPassword(id, newHashedPassword)
	if err != nil {
		fmt.Println("Error changing account password", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Sending a notification email
	if args.EnableSmtp {
		if err := h.store.SendEmail(account.Email, "Notification", "Your password has been modified"); err != nil {
			fmt.Println("Error sending email", err)
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
	}
	utilities.Reply(w, http.StatusOK, "password updated", nil, true)
}

// PUT /modify/2fa
func (h *Handlers) Change2FA(w http.ResponseWriter, r *http.Request) {
	// Getting the account ID from context
	id := utilities.GetAccountIDFromContext(r.Context())
	if id == -1 {
		fmt.Println("Error getting account ID from context")
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Getting the account by id
	account, err := h.store.GetAccountByID(id)
	if err != nil {
		fmt.Println("Error getting account by ID", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Is the account 2FA enabled?
	if account.TOTP {
		// Toggling 2FA status
		if err := h.store.ToggleAccount2FA(id, account.TOTP); err != nil {
			fmt.Println("Error changing account 2FA status", err)
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
		utilities.Reply(w, http.StatusOK, "2fa disabled", nil, true)
		return
	}
	// Generating the TOTP secret and QR code
	secret, QR, err := h.store.GenerateTOTP("squanchy Amene", account.Email, args.TotpAlgorithm)
	if err != nil {
		fmt.Println("Error generating TOTP", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Storing the secret
	if err := h.store.NewTOTP(account.ID, secret); err != nil {
		fmt.Println("Error storing TOTP secret", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Converting QR code to base64
	buffer := new(bytes.Buffer)
	png.Encode(buffer, QR)
	base64QR := base64.StdEncoding.EncodeToString(buffer.Bytes())
	utilities.Reply(w, http.StatusOK, "verify to utilize", map[string]string{
		"secret": secret,
		"QR":     base64QR,
	}, true)
}

// PUT /modify/2fa/confirm, 2FA TOTP only
func (h *Handlers) Change2FAConfirm(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.Change2FAConfirmPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusUnprocessableEntity, "invalid payload", verr, false)
		return
	}
	// Getting the account ID from context
	id := utilities.GetAccountIDFromContext(r.Context())
	if id == -1 {
		fmt.Println("Error getting account ID from context")
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Getting the account by ID
	account, err := h.store.GetAccountByID(id)
	if err != nil {
		fmt.Println("Error getting account ID", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Is the account 2FA enabled?
	if account.TOTP {
		fmt.Println("Warning someone is likely penetration testing")
		utilities.Reply(w, http.StatusUnsupportedMediaType, "2FA already enabled", nil, false)
		return
	}
	// Verifing the TOTP code
	valid, err := h.store.ValidateTOTP(id, payload.TOTP, args.TotpAlgorithm, args.TotpSkew)
	if err != nil {
		fmt.Println("Error validating TOTP", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	if !valid {
		utilities.Reply(w, http.StatusUnauthorized, "invalid TOTP code", nil, false)
		return
	}
	// Generating TOTP backup codes
	codes, err := h.store.GenerateBackupCodes(id)
	if err != nil {
		fmt.Println("Error generating TOTP backup codes", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Storing the TOTP backup codes
	if err := h.store.NewBackupCodes(id, codes); err != nil {
		fmt.Println("Error storing TOTP backup codes", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Toggling 2FA status
	var action string
	if account.TOTP {
		action = "disabled"
	} else {
		action = "enabled"
	}
	if err := h.store.ToggleAccount2FA(id, account.TOTP); err != nil {
		fmt.Println("Error changing account 2FA status", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Sending a notification email
	if args.EnableSmtp {
		if err := h.store.SendEmail(account.Email, "Notification", fmt.Sprintf("2FA %s", action)); err != nil {
			fmt.Println("Error sending email", err)
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
	}
	utilities.Reply(w, http.StatusOK, "2FA "+action, map[string][]string{
		"backup": codes,
	}, true)
}

// DELETE /delete
func (h *Handlers) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	// Parsing the JSON body
	var payload types.DeleteAccountPayload
	if err := utilities.ParseJSON(r, &payload); err != nil {
		utilities.Reply(w, http.StatusBadRequest, "status bad request", nil, false)
		return
	}
	// Validating the JSON body
	if verr := utilities.ValidateJSON(w, payload); verr != nil {
		utilities.Reply(w, http.StatusUnprocessableEntity, "invalid payload", verr, false)
		return
	}
	// Getting the account ID from context
	id := utilities.GetAccountIDFromContext(r.Context())
	if id == -1 {
		fmt.Println("Error getting account ID from context")
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Getting the account by ID
	account, err := h.store.GetAccountByID(id)
	if err != nil {
		fmt.Println("Error getting account ID", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Comparing the plaintext and hashed passwords
	if !utilities.ComparePlainAndHashedPassword(account.Password, args.Pepper, []byte(payload.Password)) {
		utilities.Reply(w, http.StatusUnauthorized, "invalid credentials", nil, false)
		return
	}
	// Invalidating all active sessions
	if err := h.store.InvalidateAllSessions(id); err != nil {
		fmt.Println("Error invalidating all sessions", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Deleting the account
	if err := h.store.DeleteAccount(id); err != nil {
		fmt.Println("Error deleting account", err)
		utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
		return
	}
	// Clearing the client authentication cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	// Sending a notification email
	if args.EnableSmtp {
		if err := h.store.SendEmail(account.Email, "Notification", "Your account has been deleted"); err != nil {
			fmt.Println("Error sending email")
			utilities.Reply(w, http.StatusInternalServerError, "internal server error", nil, false)
			return
		}
	}
	utilities.Reply(w, http.StatusOK, "account deleted", nil, true)
}

// GET /health
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(); err != nil {
		fmt.Println("Health check error", err)
		utilities.Reply(w, http.StatusServiceUnavailable, "unavailable", nil, false)
		return
	}
	utilities.Reply(w, http.StatusOK, "healthy", nil, true)
}

// squanchy /squanchy
func (h *Handlers) squanchy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Someone found the easter egg")
	utilities.Reply(w, http.StatusOK, "easter egg", nil, true)
}

// GET /openapi
func (h *Handlers) OpenAPI(w http.ResponseWriter, r *http.Request) {
	html, err := scalar.ApiReferenceHTML(&scalar.Options{
		SpecURL: fmt.Sprintf("http://localhost:6900%s/openapi/openapi.yaml", args.Prefix),
		CustomOptions: scalar.CustomOptions{
			PageTitle: "squanchy API v0.1",
		},
		Theme:              scalar.ThemeSaturn,
		DarkMode:           true,
		HideDownloadButton: true,
		Layout:             scalar.LayoutModern,
		WithDefaultFonts:   true,
	})
	if err != nil {
		fmt.Println("Error serving documentation", err)
	}
	fmt.Fprintln(w, html)
}
