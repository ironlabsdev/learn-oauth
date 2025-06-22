package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"oauth/utils/env"
	"oauth/utils/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var logFilePath = "logs/migration-app.log"

const fmtDBString = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}

	l, err := logger.New(true, logFilePath)
	if err != nil {
		panic(err)
	}

	conf := env.New()
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

	db, err := sql.Open("postgres", dbString)
	if err != nil {
		log.Fatal("Unable to connect to database ", err)
	}
	l.Info().Msg("Database connection opened successfully")

	// Create migration instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Point to your migration files. Here we're using local files, but it could be other sources.
	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	cmd := os.Args[len(os.Args)-1]
	if cmd == "up" {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			l.Fatal().
				Err(err).
				Msg("Migration up failed")
		}
		l.Info().Msg("Migration up completed successfully")
	}

	if cmd == "down" {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			l.Fatal().
				Err(err).
				Msg("Migration down failed")
		}
		l.Info().Msg("Migration down completed successfully")
	}
}
