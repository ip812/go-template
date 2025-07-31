package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form"
	"github.com/go-playground/validator/v10"
	"github.com/godruoyi/go-snowflake"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/ip812/go-template/config"
	"github.com/ip812/go-template/database"
	"github.com/ip812/go-template/logger"
	"github.com/ip812/go-template/utils"
)

const (
	dbConnectTimeout      = 10 * time.Second
	dbMaxOpenConnections  = 10
	retryMaxElapsedTime   = 15 * time.Minute
	serverIdleTimeout     = 1 * time.Minute
	serverReadTimeout     = 10 * time.Second
	serverWriteTimeout    = 30 * time.Second
	serverShutdownTimeout = 10 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.New()
	log := logger.New(cfg)

	// https://snowsta.mp
	startTime, _ := time.Parse(time.RFC3339, "2015-01-01T00:00:00Z")
	snowflake.SetStartTime(startTime)
	snowflake.SetMachineID(1)

	swappableDB := NewSwappableDB()

	server := startHTTPServer(cfg, log, swappableDB)

	db, queries, err := connectToDatabaseWithRetry(ctx, cfg, log)
	if err != nil {
		log.Error("exiting: could not connect to DB after retries: %s", err.Error())
		return
	}
	defer db.Close()

	swappableDB.Swap(db, queries)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Error("failed to set dialect: %s", err.Error())
	}
	if err := goose.Up(db, "sql/migrations"); err != nil {
		log.Error("failed to run migrations: %s", err.Error())
	}

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("error shutting down server: %s", err.Error())
	} else {
		log.Info("server shutdown cleanly")
	}
}

type dbConnection struct {
	db      *sql.DB
	queries *database.Queries
}

func connectToDatabaseWithRetry(ctx context.Context, cfg *config.Config, log logger.Logger) (*sql.DB, *database.Queries, error) {
	var conn dbConnection

	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Endpoint,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	operation := func() (dbConnection, error) {
		connCtx, cancel := context.WithTimeout(ctx, dbConnectTimeout)
		defer cancel()

		db, err := sql.Open("postgres", connectionString)
		if err != nil {
			log.Warn("failed to open the database connection: %v", err.Error())
			return conn, err
		}

		if err := db.PingContext(connCtx); err != nil {
			log.Warn("failed to ping the database: %v", err.Error())
			return conn, err
		}

		db.SetMaxOpenConns(dbMaxOpenConnections)
		log.Info("connected to database")

		conn.db = db
		conn.queries = database.New(db)
		return conn, nil
	}

	_, err := backoff.Retry[dbConnection](
		ctx,
		operation,
		backoff.WithMaxElapsedTime(retryMaxElapsedTime),
	)

	return conn.db, conn.queries, err
}

func startHTTPServer(cfg *config.Config, log logger.Logger, db DBWrapper) *http.Server {
	formDecoder := form.NewDecoder()
	formValidator := validator.New(validator.WithRequiredStructEnabled())

	handler := Handler{
		config:        cfg,
		formDecoder:   formDecoder,
		formValidator: formValidator,
		db:            db,
		log:           log,
	}

	mux := chi.NewRouter()
	mux.Handle("/static/*", handler.StaticFiles())
	mux.With().Route("/p", func(mux chi.Router) {
		mux.Route("/public", func(mux chi.Router) {
			mux.Get("/home", handler.LandingPageView)
		})
	})

	mux.Route("/api", func(mux chi.Router) {
		mux.Route("/public/v0", func(mux chi.Router) {
			mux.Route("/mailing-list", func(mux chi.Router) {
				mux.Post("/", utils.MakeTemplHandler(handler.AddEmailToMailingList))
			})
		})
	})

	mux.Get("/healthz", handler.Healthz)
	mux.NotFound(handler.LandingPageRedirect)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.Port),
		IdleTimeout:  serverIdleTimeout,
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
		Handler:      mux,
	}

	go func() {
		log.Info("server started on %s", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("cannot start server: %s", err.Error())
		}
	}()

	return server
}
