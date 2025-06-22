package pages

import (
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	db "oauth/database/generated"
	"oauth/utils/env"
	"github.com/rs/zerolog"
)

type Pages struct {
	pool    *pgxpool.Pool
	store   *sessions.CookieStore
	logger  *zerolog.Logger
	envConf *env.Conf
	queries *db.Queries
}

func NewPages(logger *zerolog.Logger, cfg *env.Conf, pool *pgxpool.Pool, store *sessions.CookieStore) *Pages {
	return &Pages{
		pool:    pool,
		store:   store,
		logger:  logger,
		envConf: cfg,
		queries: db.New(pool),
	}
}
