package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed assets/templates/dashboard.html
var templatesFS embed.FS

//go:embed assets/static
var staticFS embed.FS

// DashboardTemplate is parsed once at startup. The template is embedded and
// validated by CI, so a parse failure is a build-time bug; template.Must is the
// idiomatic way to surface it.
var DashboardTemplate = template.Must(template.ParseFS(templatesFS, "assets/templates/dashboard.html"))

// StaticHandler serves the embedded assets under /static/ (CSS, JS, vendored
// Pico). The assets live at fixed URLs but change whenever the binary is
// rebuilt, so we tell the browser to revalidate rather than caching for a fixed
// duration (which would serve stale CSS/JS after an upgrade). On localhost the
// refetch cost is negligible.
func StaticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "assets/static")
	if err != nil {
		// staticFS is embedded at build time, so this can never fail at runtime.
		panic("web: failed to scope embedded static assets: " + err.Error())
	}

	fileServer := http.StripPrefix("/static/", http.FileServerFS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		fileServer.ServeHTTP(w, r)
	})
}
