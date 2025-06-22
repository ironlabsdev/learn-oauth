package auth

import (
	"fmt"
	"net/http"

	db "oauth/database/generated"
	"oauth/utils/env"
	"oauth/web/services"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog"
)

type API struct {
	Pool        *pgxpool.Pool
	Goth        *goth.Session
	Store       *sessions.CookieStore
	Logger      *zerolog.Logger
	EnvConf     *env.Conf
	Queries     *db.Queries
	UserService *services.UserService
}

// Define a custom type for context keys to avoid collisions
type contextKey string

const (
	ProviderKey      = "provider"
	SessionID        = "session.id"
	UserIDKey        = "user.id"
	AuthenticatedKey = "authenticated"

	// Context key using the custom type
	providerContextKey contextKey = "provider"
)

func NewAuth(logger *zerolog.Logger, cfg *env.Conf, pool *pgxpool.Pool, store *sessions.CookieStore) *API {
	gothic.Store = store
	goth.UseProviders(
		google.New(cfg.GoogleAuth.ID, cfg.GoogleAuth.Secret, fmt.Sprintf("%s/auth/google/callback", cfg.GetBaseURL()), "email", "profile"),
	)

	return &API{
		Pool:        pool,
		Store:       store,
		Logger:      logger,
		EnvConf:     cfg,
		Queries:     db.New(pool),
		UserService: services.NewUserService(pool, logger),
	}
}

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	// todo - implement
}

func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	// todo - implement
}

func (a *API) Callback(w http.ResponseWriter, r *http.Request) {
	// todo - implement
}
