package store

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Availability struct {
	DoctorID     uuid.UUID `json:"doctor_id"`
	AvailableDay string    `json:"available_day"`
	StartsFrom   string    `json:"starts_from"`
	EndsAt       string    `json:"ends_at"`
}

type AvailabilityStore struct {
	db *sql.DB
}

func (s *AvailabilityStore) Create(ctx context.Context, availability *Availability) error {
	query := `
        INSERT INTO availabilities (doctor_id, available_day, starts_from, ends_at) 
        VALUES ($1, $2, $3, $4)
    `
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		availability.DoctorID,
		availability.AvailableDay,
		availability.StartsFrom,
		availability.EndsAt,
	).Scan(
		&availability.DoctorID,
		&availability.AvailableDay,
		&availability.StartsFrom,
		&availability.EndsAt,
	)
	if err != nil {
		return err
	}
	return nil

}
