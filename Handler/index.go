package Handler

import (
	"html/template"
	"net/http"
)

type IndexPageData struct {
	Username string
}

func ServeIndexPage(w http.ResponseWriter, r *http.Request) {
	username := "Utilisateur"

	cookie, err := r.Cookie("username")
	if err == nil {
		username = cookie.Value
	}

	data := IndexPageData{Username: username}

	templatePath := "./forum-B1-la-tcheam/templates/index.gohtml"

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}
