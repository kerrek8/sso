package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"os"
	"sso/internal/domain/models"
	"sso/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "postgres.New"
	db, err := sql.Open("postgres", fmt.Sprintf(storagePath, os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("DB_CONTAINER"), os.Getenv("DB_PORT")))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	fmt.Print("ping done")
	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passhash []byte) (uid int64, err error) {
	const op = "postgres.Storage.SaveUser"
	const query = `
		INSERT INTO users (email, pass_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	err = s.db.QueryRowContext(ctx, query, email, passhash).Scan(&uid)
	if err != nil {
		var postrgresErr pq.Error
		if errors.As(err, &postrgresErr) && postrgresErr.Code.Name() == "unique_violation" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)

		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return uid, nil

}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "postgres.Storage.User"
	const query = `
		SELECT id, email, pass_hash
		FROM users
		WHERE email = $1
	`

	var u models.User
	err := s.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.PassHash)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return u, nil
}

func (s *Storage) App(ctx context.Context, id int) (models.App, error) {
	const op = "postgres.Storage.App"
	const query = `SELECT id, name, secret FROM apps WHERE id = $1`
	var a models.App
	err := s.db.QueryRowContext(ctx, query, id).Scan(&a.ID, &a.Name, &a.Secret)
	if errors.Is(err, sql.ErrNoRows) {
		return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
	}
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	return a, nil
}

func (s *Storage) Close() {
	s.db.Close()
}
