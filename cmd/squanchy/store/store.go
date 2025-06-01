package store

import (
	"database/sql"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// Pings the database
func (s *Store) Ping() error {
	if err := s.db.Ping(); err != nil {
		return err
	}
	return nil
}
