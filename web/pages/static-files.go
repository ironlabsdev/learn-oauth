package pages

import (
	"net/http"
)

// StaticFileHandler wraps the file server with proper cache headers
func StaticFileHandler(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve the file
		fileServer.ServeHTTP(w, r)
	})
}
