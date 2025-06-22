package router

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	db "oauth/database/generated"
	"oauth/utils/env"
	oauthMiddleware "oauth/web/middleware"
	"oauth/web/pages"
	"oauth/web/requestlog"
	"oauth/web/requests/auth"
)

type Controller struct {
	Conf    *env.Conf
	Pool    *pgxpool.Pool
	Store   *sessions.CookieStore
	Router  *chi.Mux
	Logger  *zerolog.Logger
	Queries *db.Queries
}

func (c *Controller) RegisterUses() {
	c.Router.Use(oauthMiddleware.RequestID)
	c.Router.Use(oauthMiddleware.SetEnvConfig)
	c.Router.Use(middleware.Logger)
}

func (c *Controller) RegisterRoutes() {
	publicFS := http.Dir("public")
	c.Router.Handle("/public/*", http.StripPrefix("/public/",
		pages.StaticFileHandler(publicFS)))

	authHandler := auth.NewAuth(c.Logger, c.Conf, c.Pool, c.Store)
	c.Router.Method(http.MethodGet, "/auth/{provider}", requestlog.NewHandler(authHandler.Login, c.Logger))
	c.Router.Method(http.MethodGet, "/auth/{provider}/callback", requestlog.NewHandler(authHandler.Callback, c.Logger))
	c.Router.Method(http.MethodGet, "/auth/{provider}/logout", requestlog.NewHandler(authHandler.Logout, c.Logger))

	pageHandler := pages.NewPages(c.Logger, c.Conf, c.Pool, c.Store)
	c.Router.Method(http.MethodGet, "/", requestlog.NewHandler(pageHandler.Home, c.Logger))
	c.Router.Method(http.MethodGet, "/login", requestlog.NewHandler(pageHandler.Login, c.Logger))
	c.Router.Method(http.MethodGet, "/restricted", requestlog.NewHandler(pageHandler.Restricted, c.Logger))
}
