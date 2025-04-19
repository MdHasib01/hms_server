package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/MdHasib01/hms_server/internal/store"
	"github.com/go-chi/chi/v5"
)

type doctorKey string

const doctorCtx doctorKey = "doctor"

func (app *application) getDoctorHandler(w http.ResponseWriter, r *http.Request) {
	doctor := getDoctorFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, doctor); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) doctorContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "doctorID")
		id, err := strconv.ParseInt(idParam, 10, 64)
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

func (app *application) getDoctorsHandler(w http.ResponseWriter, r *http.Request) {
	filter := store.DoctorFilter{
		Designation:   r.URL.Query().Get("designation"),
		Qualification: r.URL.Query().Get("qualification"),
		Country:       r.URL.Query().Get("country"),
		State:         r.URL.Query().Get("state"),
		City:          r.URL.Query().Get("city"),
		Day:           r.URL.Query().Get("day"),
	}

	doctors, err := app.store.Doctors.GetDoctorsByFilter(r.Context(), filter)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, doctors); err != nil {
		app.internalServerError(w, r, err)
	}
}
