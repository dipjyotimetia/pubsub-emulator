package web

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed assets/templates/dashboard.html
var templatesFS embed.FS

//go:embed assets/static/css/dashboard.css
var cssContent embed.FS

//go:embed assets/static/js/dashboard.js
var jsContent embed.FS

var DashboardTemplate *template.Template

func init() {
	var err error
	DashboardTemplate, err = template.ParseFS(templatesFS, "assets/templates/dashboard.html")
	if err != nil {
		panic("Failed to parse dashboard template: " + err.Error())
	}
}

// ServeDashboardCSS serves the dashboard CSS file
func ServeDashboardCSS(w http.ResponseWriter, r *http.Request) {
	content, err := cssContent.ReadFile("assets/static/css/dashboard.css")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write(content)
}

// ServeDashboardJS serves the dashboard JavaScript file
func ServeDashboardJS(w http.ResponseWriter, r *http.Request) {
	content, err := jsContent.ReadFile("assets/static/js/dashboard.js")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	_, _ = w.Write(content)
}
