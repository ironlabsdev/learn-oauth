package middleware

import (
	"net/http"

	ctxUtil "oauth/utils/ctx"
	"oauth/utils/env"
)

func SetEnvConfig(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = ctxUtil.SetEnvConfigID(ctx, env.New())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
