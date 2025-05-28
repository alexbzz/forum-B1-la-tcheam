package Handler

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

func ServeIndexPage(w http.ResponseWriter, r *http.Request) {
	templatePath := filepath.Join("./templates/index.gohtml")
	fmt.Println("Tentative de charger le template:", templatePath)

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
