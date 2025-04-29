package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type Doctor struct {
	UserID         uuid.UUID `json:"user_id"`
	UserName       string    `json:"username"`
	Password       string    `json:"-"` // Typically not returned in queries
	Email          string    `json:"email"`
	FirstName      string    `json:"firstname"`
	LastName       string    `json:"lastname"`
	Age            string    `json:"age"`
	Gender         string    `json:"gender"`
	MaritalStatus  string    `json:"marital_status"`
	Designation    string    `json:"designation"`
	Qualification  string    `json:"qualification"`
	BloodGroup     string    `json:"blood_group"`
	Address        string    `json:"address"`
	Country        string    `json:"country"`
	State          string    `json:"state"`
	City           string    `json:"city"`
	PostalCode     string    `json:"postal_code"`
	Specialization string    `json:"specialization"`
	LicenseNumber  string    `json:"license_number"`
	Availability   []string  `json:"availability"`
}

type DoctorStore struct {
	db *sql.DB
}

func (s *DoctorStore) GetByID(ctx context.Context, id uuid.UUID) (*Doctor, error) {
	query := `
	SELECT 
		u.id, u.username, u.email, u.firstname, u.lastname, u.age, u.gender,
		u.marital_status, u.designation, u.qualification, u.blood_group,
		u.address, u.country, u.state, u.city, u.postal_code,
		d.specialization, d.license_number,
		COALESCE(json_agg(a.available_day) FILTER (WHERE a.available_day IS NOT NULL), '[]') AS availability
	FROM users u
	INNER JOIN doctors d ON d.user_id = u.id
	LEFT JOIN availability a ON a.doctor_id = u.id
	WHERE u.id = $1
	GROUP BY 
		u.id, u.username, u.email, u.firstname, u.lastname, u.age, u.gender,
		u.marital_status, u.designation, u.qualification, u.blood_group,
		u.address, u.country, u.state, u.city, u.postal_code,
		d.specialization, d.license_number
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var availabilityJSON []byte
	doctor := &Doctor{}

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&doctor.UserID,
		&doctor.UserName,
		&doctor.Email,
		&doctor.FirstName,
		&doctor.LastName,
		&doctor.Age,
		&doctor.Gender,
		&doctor.MaritalStatus,
		&doctor.Designation,
		&doctor.Qualification,
		&doctor.BloodGroup,
		&doctor.Address,
		&doctor.Country,
		&doctor.State,
		&doctor.City,
		&doctor.PostalCode,
		&doctor.Specialization,
		&doctor.LicenseNumber,
		&availabilityJSON,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	_ = json.Unmarshal(availabilityJSON, &doctor.Availability)

	return doctor, nil
}

func (s *DoctorStore) Create(ctx context.Context, doctor *Doctor) error {
	query := `
		
		INSERT INTO doctors (user_id,firstname, lastname, age, gender, marital_status,
			designation, qualification, blood_group, address, country, state, city, postal_code,specialization, license_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16);
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query,
		doctor.UserID,
		doctor.FirstName,
		doctor.LastName,
		doctor.Age,
		doctor.Gender,
		doctor.MaritalStatus,
		doctor.Designation,
		doctor.Qualification,
		doctor.BloodGroup,
		doctor.Address,
		doctor.Country,
		doctor.State,
		doctor.City,
		doctor.PostalCode,
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
	return err
}

func (s *DoctorStore) Delete(ctx context.Context, userID uuid.UUID) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, userID); err != nil {
			return err
		}
		return s.deleteUserInvitations(ctx, tx, userID)
	})
}

func (s *DoctorStore) delete(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	return err
}
