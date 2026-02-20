package postgres

import (
	"database/sql"
	"time"
)

const (
	connMaxLifetime time.Duration = 30 * time.Minute
	connMaxIdleTime time.Duration = 5 * time.Minute
	maxIdleConns    int           = 5
	maxOpenConns    int           = 25
	defaultTimeout  time.Duration = 5 * time.Second
)

type Postgres struct {
	db             *sql.DB
	DefaultTimeout time.Duration
}

func New(postgresURL string) (Postgres, error) {
	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		return Postgres{}, err
	}
	configureConnectionPool(db)
	return Postgres{
		db:             db,
		DefaultTimeout: defaultTimeout,
	}, nil
}

func configureConnectionPool(db *sql.DB) {
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(connMaxIdleTime)
}
