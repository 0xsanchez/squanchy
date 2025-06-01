package store

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"gopkg.in/gomail.v2"
)

type EmailConfig struct {
	Address  string
	Port     int
	Username string
	Password string
	From     string
}

// Creating a new unexported email configuration
var emailConfig *EmailConfig

func (s *Store) InitSMTP(address string, port int, username, password, from string) *EmailConfig {
	emailConfig = &EmailConfig{
		Address:  address,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
	return emailConfig
}

// Sends a verification email
func (s *Store) SendEmail(to, subject, message string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", emailConfig.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)
	d := gomail.NewDialer(emailConfig.Address, emailConfig.Port, emailConfig.Username, emailConfig.Password)
	return d.DialAndSend(m)
}

// Generates a email verification code
func (s *Store) GenerateEmailVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

// Inserts a new email verification code
func (s *Store) NewEmailVerificationCode(id int, code string) error {
	expiresAt := time.Now().Add(time.Hour)
	_, err := s.db.Exec(`INSERT INTO smtp(account, code, expires) VALUES (?, ?, ?) ON CONFLICT(account) DO UPDATE SET code = ?, expires = ?`,
		id, code, expiresAt, code, expiresAt)
	return err
}

// Deletes an email verification code
func (s *Store) DeleteEmailVerificationCodes(id int) error {
	_, err := s.db.Exec("DELETE FROM smtp WHERE account = ?", id)
	return err
}

// Compares payload and database email verification codes
func (s *Store) IsEmailVerificationCodeValid(id int, code string) (bool, error) {
	var stored string
	var expires time.Time
	err := s.db.QueryRow(`SELECT code, expires FROM smtp WHERE account = ?`, id).Scan(&stored, &expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if time.Now().After(expires) {
		return false, nil
	}
	return stored == code, nil
}
