package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func webHandler(webDir string) http.Handler {
	fileServer := http.FileServer(http.Dir(webDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/js/") && strings.HasSuffix(r.URL.Path, ".js") {
			serveJavaScript(w, r, webDir)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func serveJavaScript(w http.ResponseWriter, r *http.Request, webDir string) {
	cleanPath := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	filePath := filepath.Join(webDir, cleanPath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	body := strings.ReplaceAll(string(content), "http://localhost:3000", "")
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}
