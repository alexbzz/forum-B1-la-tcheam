package Handler

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

type IndexPageData struct {
	Username string
}

func ServeIndexPage(w http.ResponseWriter, r *http.Request) {
	templatePath := filepath.Join("./forum-B1-la-tcheam/templates/index.gohtml")
	fmt.Println("Tentative de charger le template:", templatePath)
	username := "Utilisateur"

	cookie, err := r.Cookie("username")
	if err == nil {
		username = cookie.Value
	}

	data := IndexPageData{Username: username}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}
