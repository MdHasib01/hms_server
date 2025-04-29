package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/MdHasib01/hms_server/internal/mailer"
	"github.com/MdHasib01/hms_server/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type doctorKey string

const doctorCtx doctorKey = "doctor"

// getByID godoc
//
//	@Summary		Fetches the doctor by id
//	@Description	Fetches the doctor by id
//	@Tags			doctor
//	@Accept			json
//	@Produce		json
//	@Param			doctorID	path		string	true	"Doctor ID"
//	@Success		204			{string}	string	"Doctor Found"
//	@Failure		400			{object}	error	"Doctor Id  missing"
//	@Failure		404			{object}	error	"Dctor not found"
//	@Security		ApiKeyAuth
//	@Router			/doctors/{doctorID} [get]
func (app *application) GetByID(w http.ResponseWriter, r *http.Request) {
	doctor := getDoctorFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, doctor); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) doctorContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "doctorID")
		fmt.Print(idParam)
		id, err := uuid.Parse(idParam)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		ctx := r.Context()

		doctor, err := app.store.Doctors.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, doctorCtx, doctor)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getDoctorFromCtx(r *http.Request) *store.Doctor {
	doctor, _ := r.Context().Value(doctorCtx).(*store.Doctor)
	return doctor
}

type CreateDoctorPayload struct {
	Username       string `json:"username" validate:"required,max=100"`
	Email          string `json:"email" validate:"required,email,max=255"`
	Password       string `json:"password" validate:"required,min=3,max=72"`
	FirstName      string `json:"firstname" validate:"required,max=100"`
	LastName       string `json:"lastname" validate:"required,max=100"`
	Age            string `json:"age" validate:"required"`
	Gender         string `json:"gender" validate:"required"`
	MaritalStatus  string `json:"marital_status" validate:"required"`
	Designation    string `json:"designation" validate:"required"`
	Qualification  string `json:"qualification" validate:"required"`
	BloodGroup     string `json:"blood_group" validate:"required"`
	Address        string `json:"address" validate:"required"`
	Country        string `json:"country" validate:"required"`
	State          string `json:"state" validate:"required"`
	City           string `json:"city" validate:"required"`
	PostalCode     string `json:"postal_code" validate:"required"`
	Specialization string `json:"specialization" validate:"required"`
	LicenseNumber  string `json:"license_number" validate:"required"`
}

// CreateDoctorHandler godoc
//
//	@Summary		Creates a doctor account
//	@Description	Creates a new doctor user and doctor profile
//	@Tags			doctor
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateDoctorPayload	true	"Doctor user and profile information"
//	@Success		201		{object}	store.Doctor
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/doctors [post]
func (app *application) CreateDoctorHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateDoctorPayload

	// Step 1: Read and validate input
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	// Step 2: Create User object
	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	// Hash the password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Step 3: Generate activation token
	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	// Step 4: Save user in DB with invitation
	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail, store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// Step 5: Create Doctor profile
	doctor := &store.Doctor{
		UserID:         user.ID,
		UserName:       user.Username,
		Email:          user.Email,
		FirstName:      payload.FirstName,
		LastName:       payload.LastName,
		Age:            payload.Age,
		Gender:         payload.Gender,
		MaritalStatus:  payload.MaritalStatus,
		Designation:    payload.Designation,
		Qualification:  payload.Qualification,
		BloodGroup:     payload.BloodGroup,
		Address:        payload.Address,
		Country:        payload.Country,
		State:          payload.State,
		City:           payload.City,
		PostalCode:     payload.PostalCode,
		Specialization: payload.Specialization,
		LicenseNumber:  payload.LicenseNumber,
	}

	err = app.store.Doctors.Create(ctx, doctor)
	if err != nil {
		// Rollback user if doctor creation fails
		_ = app.store.Users.Delete(ctx, user.ID)
		app.internalServerError(w, r, err)
		return
	}

	// Step 6: Send welcome email
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)
	isProdEnv := app.config.env == "production"

	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	status, err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)
		_ = app.store.Doctors.Delete(ctx, doctor.UserID)
		_ = app.store.Users.Delete(ctx, user.ID)
		app.internalServerError(w, r, err)
		return
	}

	app.logger.Infow("Email sent", "status", status)

	// Step 7: Return Full Doctor Profile Response (excluding password)
	resp := struct {
		UserID         uuid.UUID `json:"user_id"`
		UserName       string    `json:"username"`
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
	}{
		UserID:         doctor.UserID,
		UserName:       doctor.UserName,
		Email:          doctor.Email,
		FirstName:      doctor.FirstName,
		LastName:       doctor.LastName,
		Age:            doctor.Age,
		Gender:         doctor.Gender,
		MaritalStatus:  doctor.MaritalStatus,
		Designation:    doctor.Designation,
		Qualification:  doctor.Qualification,
		BloodGroup:     doctor.BloodGroup,
		Address:        doctor.Address,
		Country:        doctor.Country,
		State:          doctor.State,
		City:           doctor.City,
		PostalCode:     doctor.PostalCode,
		Specialization: doctor.Specialization,
		LicenseNumber:  doctor.LicenseNumber,
	}

	if err := app.jsonResponse(w, http.StatusCreated, resp); err != nil {
		app.internalServerError(w, r, err)
	}
}
