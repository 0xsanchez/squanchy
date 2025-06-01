package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/0xsanchez/squanchy/cmd/squanchy/types"
	_ "modernc.org/sqlite"
)

// Queries the database for an account with a given id
func (s *Store) GetAccountByID(id int) (*types.Account, error) {
	rows, err := s.db.Query("SELECT * FROM account WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	a := new(types.Account)
	for rows.Next() {
		a, err = scanRowIntoAccount(rows)
		if err != nil {
			return nil, err
		}
	}
	// Checking if an account has been found
	if a.ID == 0 {
		return nil, fmt.Errorf("account not found")
	}
	return a, nil
}

// Queries the database for an account with a given email
func (s *Store) GetAccountByEmail(email string) (*types.Account, error) {
	rows, err := s.db.Query("SELECT * FROM account WHERE email = ?", email)
	if err != nil {
		return nil, err
	}
	a := new(types.Account)
	for rows.Next() {
		a, err = scanRowIntoAccount(rows)
		if err != nil {
			return nil, err
		}
	}
	// Checking if an account has been found
	if a.ID == 0 {
		return nil, fmt.Errorf("account not found")
	}
	return a, nil
}

// Scans account rows
func scanRowIntoAccount(rows *sql.Rows) (*types.Account, error) {
	account := new(types.Account)
	err := rows.Scan(
		&account.ID,
		&account.Email,
		&account.Password,
		&account.LastLogin,
		&account.FailedAttempts,
		&account.LockedUntil,
		&account.Verified,
		&account.TOTP,
		&account.Updated,
		&account.Created,
	)
	if err != nil {
		return nil, err
	}
	return account, nil
}

// Inserts a new account in the database
func (s *Store) NewAccount(email, hashedPassword string) error {
	_, err := s.db.Exec("INSERT INTO account (email, password) VALUES (?, ?)", email, hashedPassword)
	return err
}

// Updates an account email address
func (s *Store) ChangeAccountEmail(id int, newEmail string) error {
	result, err := s.db.Exec("UPDATE account SET email = ?, updated = CURRENT_TIMESTAMP WHERE id = ?", newEmail, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

// Changes an account password
func (s *Store) ChangeAccountPassword(id int, newPassword string) error {
	_, err := s.db.Exec(`UPDATE account SET password = ?, updated = CURRENT_TIMESTAMP, attempts = 0, locked = CURRENT_TIMESTAMP WHERE id = ?`, newPassword, id)
	return err
}

// Changes an account 2FA status
func (s *Store) ToggleAccount2FA(id int, value bool) error {
	_, err := s.db.Exec(`UPDATE account SET totp = ?, updated = CURRENT_TIMESTAMP WHERE id = ?`, !value, id)
	return err
}

// Changes an account verification status
func (s *Store) VerifyAccount(id int) error {
	_, err := s.db.Exec(`UPDATE account SET verified = TRUE, updated = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}

// Deletes the account by id
func (s *Store) DeleteAccount(id int) error {
	// Use transaction for atomic deletion
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Delete all account related data across tables
	if _, err := tx.Exec("DELETE FROM jwt WHERE account = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM totp WHERE account = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM totp_backup WHERE account = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM account WHERE id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

// Changes an account login attempts
func (s *Store) TrackLoginAttempts(id, attempts int) error {
	_, err := s.db.Exec(`UPDATE account SET attempts = ?, updated = CURRENT_TIMESTAMP WHERE id = ?`, attempts, id)
	return err
}

// Lockes an account
func (s *Store) LockAccount(id int, lockTime time.Time) error {
	// Formatting locking time
	formatted := lockTime.Format(time.RFC3339)
	_, err := s.db.Exec(`UPDATE account SET locked = ?, updated = CURRENT_TIMESTAMP WHERE id = ?`, formatted, id)
	return err
}

// Resets an account login attempts
func (s *Store) ResetLoginAttempts(id int) error {
	_, err := s.db.Exec(`UPDATE account SET attempts = 0, locked = CURRENT_TIMESTAMP, last = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}

// Changes an account last login
func (s *Store) ChangeLastLogin(id int) error {
	_, err := s.db.Exec(`UPDATE account SET last = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}
