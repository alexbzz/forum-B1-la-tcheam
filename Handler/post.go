package Handler

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

var (
	store *sessions.CookieStore
	db    *sql.DB
)

func SetStore(s *sessions.CookieStore) {
	store = s
}

func SetDB(database *sql.DB) {
	db = database
}

func ServePostPage(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	username := session.Values["username"]

	data := struct {
		Username interface{}
	}{
		Username: username,
	}

	tmpl, err := template.ParseFiles("templates/post.gohtml")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session, _ := store.Get(r, "session-name")
	username := session.Values["username"]
	if username == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	category := r.FormValue("category")

	_, err := db.Exec(`
        INSERT INTO posts (title, content, category, author, created_at)
        VALUES (?, ?, ?, ?, ?)
    `, title, content, category, username, time.Now())

	if err != nil {
		http.Error(w, "Erreur lors de la création du post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
