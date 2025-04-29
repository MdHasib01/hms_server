package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Doctors interface {
		GetByID(context.Context, uuid.UUID) (*Doctor, error)
		Create(context.Context, *Doctor) error
		Delete(context.Context, uuid.UUID) error
		GetAllDoctors(context.Context) ([]*Doctor, error)
	}
	Users interface {
		GetByID(context.Context, uuid.UUID) (*User, error)
		GetByEmail(context.Context, string) (*User, error)
		Create(context.Context, *sql.Tx, *User) error
		CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error
		Activate(context.Context, string) error
		Delete(context.Context, uuid.UUID) error
		CreateWithRole(context.Context, *User, int) error
	}

	Availability interface {
		Create(context.Context, *Availability) error
	}
	Roles interface {
		GetByName(context.Context, string) (*Role, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Doctors: &DoctorStore{db},
		Users:   &UserStore{db},
		Roles:   &RoleStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
