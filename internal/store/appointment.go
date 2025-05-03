package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Appointment struct {
	ID              uuid.UUID `json:"id"`
	DoctorID        uuid.UUID `json:"doctor_id"`
	PatientID       uuid.UUID `json:"patient_id"`
	AppointmentTime time.Time `json:"appointment_time"`
	DoctorEmail     string    `json:"doctor_first_name"`
	PatientEmail    string    `json:"patient_email"`
}

type AppointmentStore struct {
	db *sql.DB
}

func NewAppointmentStore(db *sql.DB) *AppointmentStore {
	return &AppointmentStore{db: db}
}

func (s *AppointmentStore) Create(ctx context.Context, appointment *Appointment) error {
	query := `
		INSERT INTO appointment (doctor_id, patient_id, appointment_time)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query,
		appointment.DoctorID,
		appointment.PatientID,
		appointment.AppointmentTime,
	).Scan(&appointment.ID)

	return err
}

func (s *AppointmentStore) GetAllAppointments(ctx context.Context) ([]*Appointment, error) {
	query := `
		SELECT
			a.id,
			a.doctor_id,
			a.patient_id,
			a.appointment_time,
			u_patient.email AS patient_email,
			u_doctor.email AS doctor_email
		FROM appointment a
		JOIN users u_patient ON a.patient_id = u_patient.id
		JOIN doctors d ON a.doctor_id = d.user_id
		JOIN users u_doctor ON d.user_id = u_doctor.id;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appointments := []*Appointment{}

	for rows.Next() {
		appointment := &Appointment{}
		err := rows.Scan(
			&appointment.ID,
			&appointment.DoctorID,
			&appointment.PatientID,
			&appointment.AppointmentTime,
			&appointment.PatientEmail,
			&appointment.DoctorEmail,
		)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, appointment)
	}

	return appointments, nil
}
