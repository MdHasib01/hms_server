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
	Specialization string `json:"specialization" validate:"required,max=100"`
	LicenseNumber  string `json:"license_number" validate:"required,max=50"`
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
	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp) // 2 = doctor role
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

		// rollback both doctor + user if email fails
		_ = app.store.Doctors.Delete(ctx, doctor.UserID)
		_ = app.store.Users.Delete(ctx, user.ID)

		app.internalServerError(w, r, err)
		return
	}

	app.logger.Infow("Email sent", "status", status)

	// Step 7: Return Success Response
	resp := struct {
		ID             uuid.UUID `json:"id"`
		Username       string    `json:"username"`
		Email          string    `json:"email"`
		Specialization string    `json:"specialization"`
		LicenseNumber  string    `json:"license_number"`
	}{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Specialization: doctor.Specialization,
		LicenseNumber:  doctor.LicenseNumber,
	}

	if err := app.jsonResponse(w, http.StatusCreated, resp); err != nil {
		app.internalServerError(w, r, err)
	}
}
