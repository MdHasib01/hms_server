package main

import (
	"fmt"
	"net/http"

	"github.com/MdHasib01/hms_server/internal/store"
	"github.com/google/uuid"
)

type CreateAvailabilityPayload struct {
	DoctorID     uuid.UUID `json:"doctor_id"`
	AvailableDay string    `json:"available_day"`
	StartsFrom   string    `json:"starts_from"`
	EndsAt       string    `json:"ends_at"`
}

// CreateAvailabilityHandler godoc
//
//	@Summary		Creates a new availability entry
//	@Description	Creates an availability entry for a doctor
//	@Tags			doctor
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateAvailabilityPayload	true	"Availability details"
//	@Success		201		{object}	store.Availability
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/doctors/availability [post]
func (app *application) CreateAvailablityHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateAvailabilityPayload

	// Step 1: Read and validate input
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// 🛑 Step 2: Validate DoctorID is a real UUID
	doctorUUID, err := uuid.Parse(payload.DoctorID)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid doctor_id format: %v", err))
		return
	}

	ctx := r.Context()

	// Step 3: Create Availability entry
	availability := &store.Availability{
		DoctorID:     doctorUUID,
		AvailableDay: payload.AvailableDay,
		StartsFrom:   payload.StartsFrom,
		EndsAt:       payload.EndsAt,
	}

	err = app.store.Availability.Create(ctx, availability)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
