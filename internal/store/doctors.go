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
func (s *DoctorStore) Create(ctx context.Context, doctor *Doctor) error {
	query := `
		INSERT INTO doctors (user_id, specialization, license_number) 
		VALUES ($1, $2, $3)
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query,
		doctor.UserID,
		doctor.Specialization,
		doctor.LicenseNumber,
	)

	return err
}
func (s *DoctorStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}
func (s *DoctorStore) Delete(ctx context.Context, userID uuid.UUID) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, userID); err != nil {
			return err
		}

		if err := s.deleteUserInvitations(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})
}

func (s *DoctorStore) delete(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
