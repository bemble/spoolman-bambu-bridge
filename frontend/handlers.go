package frontend

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

// RegisterRoutes registers frontend routes on the given mux.
func RegisterRoutes(mux *http.ServeMux) {
	tmpl, err := template.ParseFS(Assets, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	staticFS, _ := fs.Sub(Assets, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "dashboard.html", nil)
	})
}
