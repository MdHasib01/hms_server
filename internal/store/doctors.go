package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type Doctor struct {
	UserID         uuid.UUID `json:"user_id"`
	UserName       string    `json:"username"`
	Password       string    `json:"-"`
	Email          string    `json:"email"`
	Specialization string    `json:"specialization"`
	LicenseNumber  string    `json:"license_number"`
}

type DoctorStore struct {
	db *sql.DB
}

func (s *DoctorStore) GetByID(ctx context.Context, id uuid.UUID) (*Doctor, error) {
	query := `
       SELECT d.user_id, u.username, u.email, d.specialization, d.license_number
FROM doctors d
JOIN users u ON d.user_id = u.id
WHERE d.user_id = $1;
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var doctor = &Doctor{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&doctor.UserID,
		&doctor.UserName,
		&doctor.Email,
		&doctor.Specialization,
		&doctor.LicenseNumber,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return doctor, nil
}
