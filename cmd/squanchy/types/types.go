package types

import "time"

// ---------------------------------------------
// Account database model and request payloads |
// ---------------------------------------------

type Account struct {
	ID             int       `db:"id"`
	Email          string    `db:"email"`
	Password       string    `db:"password"`
	FailedAttempts int       `db:"attempts"`
	LockedUntil    time.Time `db:"locked"`
	Verified       bool      `db:"verified"`
	TOTP           bool      `db:"totp"`
	LastLogin      time.Time `db:"last"`
	Updated        time.Time `db:"updated"`
	Created        time.Time `db:"created"`
}

type RegisterPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,passwordpolicy"`
}

type RegisterResendPayload struct {
	Email string `json:"email" validate:"required,email"`
}

type RegisterConfirmPayload struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,numeric,len=6"`
}

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,passwordpolicy"`
	TOTP     string `json:"totp,omitempty" validate:"omitempty,numeric,len=6"`
}

type RecoveryPayload struct {
	Email string `json:"email" validate:"required,email"`
}

type RecoveryConfirmPayload struct {
	New   string `json:"new" validate:"required,passwordpolicy"`
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,numeric,len=6"`
}

type RecoveryTOTPPayload struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,alphanum,uppercase,len=8"`
}

type ChangeEmailPayload struct {
	New string `json:"new" validate:"required,email"`
}

type ChangePasswordPayload struct {
	New      string `json:"new" validate:"required,passwordpolicy"`
	Password string `json:"password" validate:"required,passwordpolicy"`
}

type Change2FAConfirmPayload struct {
	TOTP string `json:"totp" validate:"required,numeric,len=6"`
}

type DeleteAccountPayload struct {
	Password string `json:"password" validate:"required,passwordpolicy"`
}
