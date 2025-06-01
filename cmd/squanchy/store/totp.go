package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"image"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var mapping = map[string]otp.Algorithm{
	"1":   otp.AlgorithmSHA1,
	"256": otp.AlgorithmSHA256,
	"512": otp.AlgorithmSHA512,
}

// Genereates a new TOTP secret key and QR code
func (s *Store) GenerateTOTP(issuer, email, algorithm string) (secret string, qr image.Image, err error) {
	// Getting the TOTP algorithm
	strenght, ok := mapping[algorithm]
	if !ok {
		return "", nil, fmt.Errorf("invalid algorithm specified")
	}
	// Generate random secret
	secretBytes := make([]byte, 20)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", nil, err
	}
	secret = base32.StdEncoding.EncodeToString(secretBytes)
	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
		Secret:      []byte(secret),
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   strenght,
	})
	if err != nil {
		return "", nil, err
	}
	// Generate QR code
	qr, err = key.Image(200, 200)
	if err != nil {
		return "", nil, err
	}
	return secret, qr, nil
}

// Inserts a TOTP secret in the database(can replace existing)
func (s *Store) NewTOTP(id int, secret string) error {
	_, err := s.db.Exec(`INSERT INTO totp (account, secret) VALUES (?, ?) ON CONFLICT(account) DO UPDATE SET secret = ?`, id, secret, secret)
	return err
}

// Generates a new set of backup codes
func (s *Store) GenerateBackupCodes(id int) ([]string, error) {
	codes := make([]string, 8)
	for i := 0; i < len(codes); i++ {
		codeBytes := make([]byte, 8)
		if _, err := rand.Read(codeBytes); err != nil {
			return []string{}, err
		}
		codes[i] = base32.StdEncoding.EncodeToString(codeBytes)[:8]
	}
	return codes, nil
}

// Insert TOTP backup codes in the database
func (s *Store) NewBackupCodes(id int, codes []string) error {
	// Begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	// Deleting existing backup codes
	_, err = tx.Exec("DELETE FROM totp_backup WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Insert new ones
	for _, code := range codes {
		_, err = tx.Exec("INSERT INTO totp_backup (account, code) VALUES (?, ?)", id, code)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// Validates a TOTP code using the database stored secret
func (s *Store) ValidateTOTP(id int, code, algorithm string, skew uint) (bool, error) {
	// Getting the TOTP algorithm
	strenght, ok := mapping[algorithm]
	if !ok {
		return false, fmt.Errorf("invalid algorithm specified")
	}
	var secret string
	if err := s.db.QueryRow("SELECT secret FROM totp WHERE account = ?", id).Scan(&secret); err != nil {
		return false, err
	}
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      skew,
		Digits:    otp.DigitsSix,
		Algorithm: strenght,
	})
	if err != nil {
		return false, err
	}
	if !valid {
		return false, nil
	}
	return true, nil
}

// Validates a TOTP backup code and consumes it
func (s *Store) IsTOTPBackupCodeValid(id int, code string) (bool, error) {
	var stored string
	if err := s.db.QueryRow("SELECT code FROM totp_backup WHERE account = ? AND code = ?", id, code).Scan(&stored); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	// Deleting used codes
	if _, err := s.db.Exec("DELETE FROM totp_backup WHERE account = ? AND code = ?", id, code); err != nil {
		return false, err
	}
	return stored == code, nil
}
