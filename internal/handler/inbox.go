package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"forum-B1-la-tcheam/internal/models"
)

type InboxPageData struct {
	Username string
	Messages []models.PrivateMessage
}

func InboxHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		username := r.Context().Value("username").(string)

		messages, err := models.GetInboxMessages(db, userID)
		if err != nil {
			http.Error(w, "Erreur de chargement des messages", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("templates/inbox.html")
		if err != nil {
			http.Error(w, "Erreur de template", http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, InboxPageData{
			Username: username,
			Messages: messages,
		})
	}
}
