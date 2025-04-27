package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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
//	@Param			doctorID	path		int		true	"Doctor ID"
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
