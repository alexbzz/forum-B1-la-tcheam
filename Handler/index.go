package Handler

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func ServeIndexPage(w http.ResponseWriter, r *http.Request) {
	// Chemin direct vers votre dossier templates
	templatePath := filepath.Join("C:\\Users\\raphy\\GolandProjects\\forum-B1-la-tcheam\\templates", "index.html")

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
