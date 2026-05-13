package handler

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type WebHandler struct {
	templates *template.Template
}

func NewWebHandler() *WebHandler {
	// Загрузка HTML шаблонов
	tmpl := template.Must(template.ParseFiles(
		filepath.Join("web", "index.html"),
		filepath.Join("web", "dashboard.html"),
	))

	return &WebHandler{
		templates: tmpl,
	}
}

// Главная страница
func (h *WebHandler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "index.html", nil)
}

// Страница дашборда
func (h *WebHandler) ServeDashboard(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "dashboard.html", nil)
}

// Статические файлы (CSS, JS)
func (h *WebHandler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", r.URL.Path))
}
