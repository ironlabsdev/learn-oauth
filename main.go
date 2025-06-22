package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	chi "github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	db "oauth/database/generated"
	"oauth/utils/env"
	"oauth/utils/logger"
	ironRouter "oauth/web/router"
)

var logFilePath = "logs/app.log"
const fmtDBString = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found")
	}
	conf := env.New()
	l, err := logger.New(conf.Server.Debug, logFilePath)
	if err != nil {
		log.Fatalf("Could not create log file")
	}

	l.Info().Str("host", conf.DB.Host).Int("port", conf.DB.Port).Str("database", conf.DB.DBName).Msg("Connecting to database")
	dbString := fmt.Sprintf(fmtDBString, conf.DB.Host, conf.DB.Username, conf.DB.Password, conf.DB.DBName, conf.DB.Port)
	dbConfig, err := pgxpool.ParseConfig(dbString)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to parse database config")
		return
	}
	pool, err := pgxpool.NewWithConfig(
		context.Background(),
		dbConfig,
	)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to create database connection pool")
		return
	}
	defer pool.Close()
	l.Info().Msg("Database connection established successfully")
	queries := db.New(pool)
	store := createCookieStore(conf)
	chiRouter := chi.NewRouter()

	routerController := ironRouter.Controller{
		Pool:    pool,
		Conf:    conf,
		Store:   store,
		Logger:  l,
		Router:  chiRouter,
		Queries: queries,
	}

	routerController.RegisterRoutes()

	addr := fmt.Sprintf("0.0.0.0:%d", conf.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      routerController.Router,
		ReadTimeout:  conf.Server.TimeoutRead,
		WriteTimeout: conf.Server.TimeoutWrite,
		IdleTimeout:  conf.Server.TimeoutIdle,
	}

	closed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		l.Info().Msgf("Shutting down server %v", server.Addr)

		ctx, cancel := context.WithTimeout(context.Background(), conf.Server.TimeoutIdle)
		defer cancel()

		if err = server.Shutdown(ctx); err != nil {
			l.Error().Err(err).Msg("Server shutdown failure")
		}

		if err == nil {
			pool.Close()
			l.Info().Msg("DB connection closed")
		}

		close(closed)
	}()

	l.Info().Str("address", addr).Int("port", conf.Server.Port).Msg("Starting HTTP server")
	if serverCloseErr := server.ListenAndServe(); serverCloseErr != nil && !errors.Is(serverCloseErr, http.ErrServerClosed) {
		l.Fatal().Err(serverCloseErr).Msg("Server startup failure")
	}

	<-closed
	l.Info().Msgf("Server shutdown successfully")
}

func createCookieStore(conf *env.Conf) *sessions.CookieStore {
	store := sessions.NewCookieStore(conf.Server.Secret)
	store.MaxAge(conf.Auth.MaxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.SameSite = http.SameSiteLaxMode

	return store
}
