package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// spaHandler serves the built React app from dir. Real files are served
// directly; any other (non-API) path falls back to index.html so client-side
// routing works on refresh/deep links.
func spaHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	index := filepath.Join(dir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never let the SPA shadow the API.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// If the requested file exists, serve it; otherwise serve index.html.
		clean := filepath.Clean(r.URL.Path)
		if clean == "/" {
			http.ServeFile(w, r, index)
			return
		}
		if info, err := os.Stat(filepath.Join(dir, clean)); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, index)
	})
}

// dirExists reports whether path is an existing directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
