package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form"
	"github.com/go-playground/validator/v10"
	"github.com/godruoyi/go-snowflake"
	"github.com/ip812/go-template/config"
	"github.com/ip812/go-template/database"
	"github.com/ip812/go-template/logger"
	"github.com/ip812/go-template/utils"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.New()
	log := logger.New(cfg)

	// https://snowsta.mp
	snowflake.SetStartTime(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC))
	snowflake.SetMachineID(1)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Endpoint,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := connectWithBackoff(ctx, dbUrl, log, 15*time.Minute)
	if err != nil {
		log.Error("could not connect to DB: %s", err.Error())
		return
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Error("failed to set dialect: %s", err.Error())
		return
	}
	if err := goose.Up(db, "sql/migrations"); err != nil {
		log.Error("failed to run migrations: %s", err.Error())
		return
	}

	db.SetMaxOpenConns(10)
	queries := database.New(db)

	formDecoder := form.NewDecoder()
	formValidator := validator.New(validator.WithRequiredStructEnabled())

	hnd := Handler{
		config:        cfg,
		formDecoder:   formDecoder,
		formValidator: formValidator,
		db:            db,
		queries:       queries,
		log:           log,
	}

	mux := chi.NewRouter()
	mux.Handle("/static/*", hnd.StaticFiles())
	mux.With().Route("/p", func(mux chi.Router) {
		mux.Route("/public", func(mux chi.Router) {
			mux.Get("/home", hnd.HomeView)
			mux.Get("/login", hnd.LoginView)
		})
		mux.Route("/client", func(mux chi.Router) {})
		mux.Route("/admin", func(mux chi.Router) {})
	})
	mux.Route("/api", func(mux chi.Router) {
		mux.Route("/public/v0", func(mux chi.Router) {
			mux.Route("/mailing-list", func(mux chi.Router) {
				mux.Post("/", utils.MakeTemplHandler(hnd.AddEmailToMailingList))
			})
		})
		mux.Route("/client/v0", func(mux chi.Router) {})
		mux.Route("/admin/v0", func(mux chi.Router) {})
	})
	mux.Get("/healthz", hnd.Healthz)
	mux.NotFound(hnd.HomeRedirect)

	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Info("server started on %s", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error: %s", err.Error())
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)
	log.Info("server shut down")
}

func connectWithBackoff(ctx context.Context, dbUrl string, log logger.Logger, timeout time.Duration) (*sql.DB, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	bo := backoff.WithContext(backoff.NewExponentialBackOff(), timeoutCtx)

	var db *sql.DB
	op := func() error {
		var err error
		db, err = sql.Open("postgres", dbUrl)
		if err != nil {
			log.Warn("sql.Open failed: %s", err.Error())
			return err
		}
		if pingErr := db.PingContext(timeoutCtx); pingErr != nil {
			log.Warn("DB ping failed: %s", pingErr.Error())
			db.Close()
			return pingErr
		}
		log.Info("Connected to database")
		return nil
	}

	if err := backoff.Retry(op, bo); err != nil {
		return nil, err
	}
	return db, nil
}
