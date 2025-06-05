package Handler

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func ServeIndexPage(w http.ResponseWriter, r *http.Request) {
	templatePath := filepath.Join("templates/index.html")

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
