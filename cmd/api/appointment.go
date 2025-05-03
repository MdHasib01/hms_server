package main

import (
	"net/http"
	"time"

	"github.com/MdHasib01/hms_server/internal/store"
	"github.com/google/uuid"
)

type appointmentKey string

const appointmentCtx appointmentKey = "appointment"

// CreateAppointmentPayload defines the expected request body
type CreateAppointmentPayload struct {
	PatientID       uuid.UUID `json:"patient_id" validate:"required"`
	DoctorID        uuid.UUID `json:"doctor_id" validate:"required"`
	AppointmentTime time.Time `json:"appointment_time" validate:"required"`
}

// CreateAppointmentHandler godoc
//
//	@Summary		Create new appointment
//	@Description	Creates a new appointment
//	@Tags			appointment
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateAppointmentPayload	true	"Appointment Details"
//	@Success		201		{object}	store.Appointment
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/appointments [post]
func (app *application) CreateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateAppointmentPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	appointment := &store.Appointment{
		PatientID:       payload.PatientID,
		DoctorID:        payload.DoctorID,
		AppointmentTime: payload.AppointmentTime,
	}

	err := app.store.Appointments.Create(r.Context(), appointment)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, appointment); err != nil {
		app.internalServerError(w, r, err)
	}
}

// GetAllAppointmentsHandler godoc
//
//	@Summary		Get all appointments with patient and doctor info
//	@Description	Retrieves all appointments, showing patient and doctor names with appointment times
//	@Tags			appointment
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		store.Appointment
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/appointments [get]
func (app *application) GetAllAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	appointments, err := app.store.Appointments.GetAllAppointments(ctx)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, appointments); err != nil {
		app.internalServerError(w, r, err)
	}
}
