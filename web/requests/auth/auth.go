package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	db "oauth/database/generated"
	"oauth/utils/env"
	"oauth/web/services"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgtype"
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
	provider := chi.URLParam(r, ProviderKey)
	r = r.WithContext(context.WithValue(r.Context(), providerContextKey, provider))
	a.Logger.Info().Str("provider", provider).Msg("New login provider request")

	session, _ := a.Store.Get(r, SessionID)
	for key, val := range session.Values {
		a.Logger.Info().Interface("key", key).Msg("Login provider get cookie storage")
		a.Logger.Info().Interface("session_val", val).Msg("Login provider get cookie storage")
	}

	isAuthenticated := session.Values[AuthenticatedKey]
	a.Logger.Info().Interface("isAuthenticated", isAuthenticated).Msg("Check if user is already authenticated")

	if isAuthenticated != nil && isAuthenticated == true {
		a.Logger.Info().Msg("User is already authenticated. Redirecting to base url")

		http.Redirect(w, r, a.EnvConf.GetBaseURL(), http.StatusFound)
	}

	// try to get the user without re-authenticating
	if user, err := gothic.CompleteUserAuth(w, r); err == nil {
		a.Logger.Info().Interface("user", user).Msg("Got user information")
		dbUser, err := a.Queries.GetUserByEmail(r.Context(), pgtype.Text{
			String: user.Email,
			Valid:  true,
		})
		if err != nil {
			a.Logger.Err(err).Str("email", user.Email).Msg("Could not find user by email")
			http.Redirect(w, r, a.EnvConf.GetBaseURL(), http.StatusInternalServerError)
			return
		}
		session.Values[AuthenticatedKey] = true
		session.Values[UserIDKey] = dbUser.ID.String()
		err = session.Save(r, w)
		if err != nil {
			a.Logger.Err(err).Msg("Could not save session")
			http.Redirect(w, r, a.EnvConf.GetBaseURL(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, a.EnvConf.GetBaseURL(), http.StatusFound)
		return
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, ProviderKey)
	r = r.WithContext(context.WithValue(r.Context(), providerContextKey, provider))

	session, _ := a.Store.Get(r, SessionID)
	session.Values[AuthenticatedKey] = false
	session.Values[UserIDKey] = nil
	err := session.Save(r, w)
	if err != nil {
		a.Logger.Err(err).Msg("Could not save session")
		http.Redirect(w, r, a.EnvConf.GetBaseURL(), http.StatusInternalServerError)
		return
	}

	err = gothic.Logout(w, r)
	if err != nil {
		a.Logger.Err(err).Msg("Could not logout user")
		return
	}

	a.Logger.Info().Msg("User successfully logged out")
	http.Redirect(w, r, a.EnvConf.GetBaseURL(), http.StatusFound)
}

func (a *API) Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, ProviderKey)
	r = r.WithContext(context.WithValue(r.Context(), providerContextKey, provider))

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		a.Logger.Err(err).Msg("Error completing user authentication with gothic.")
		return
	}
	req := a.UserService.CreateUserFromGothUser(user)

	// Hash the user ID for security
	hasher := sha256.New()
	hasher.Write([]byte(user.IDToken))
	req.IDToken = []byte(hex.EncodeToString(hasher.Sum(nil)))

	userWithOAuth, err := a.UserService.CreateUserWithOAuth(r.Context(), req)
	if err != nil {
		a.Logger.Err(err).Msg("Failed to create user with OAuth")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	a.Logger.Info().
		Interface("user", userWithOAuth.User).
		Interface("oauth", userWithOAuth.OAuthIdentity).
		Msg("User created/updated successfully")

	session, _ := a.Store.Get(r, SessionID)
	session.Values[AuthenticatedKey] = true
	session.Values[UserIDKey] = userWithOAuth.User.ID.String()
	err = session.Save(r, w)
	if err != nil {
		a.Logger.Err(err).Msg("Could not save session")
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, a.EnvConf.GetBaseURL(), http.StatusFound)
}
