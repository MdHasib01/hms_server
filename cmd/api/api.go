package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/MdHasib01/hms_server/docs"
	"github.com/MdHasib01/hms_server/internal/auth"
	"github.com/MdHasib01/hms_server/internal/mailer"
	"github.com/MdHasib01/hms_server/internal/store"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	auth        authConfig
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type basicConfig struct {
	user string
	pass string
}

type mailConfig struct {
	// sendGrid  sendGridConfig
	mailTrap mailTrapConfig

	fromEmail string
	exp       time.Duration
}

// type sendGridConfig struct {
// 	apiKey string
// }

type mailTrapConfig struct {
	apiKey string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)

				r.Get("/", app.getUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
			})
		})

		// Doctor routes
		r.Route("/doctors", func(r chi.Router) {
			// r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.CreateDoctorHandler)
			r.Get("/", app.getAllDoctorsHandler)
			r.Route("/{doctorID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Use(app.doctorContextMiddleware)
				r.Get("/", app.GetByID)
			})
			r.Route("/availability", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Post("/", app.activateUserHandler)
			})

		})

		// Public routes
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)

	return srv.ListenAndServe()
}
