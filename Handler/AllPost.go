package Handler

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type PostView struct {
	ID           int
	Title        string
	Content      string
	Author       string
	Likes        int
	Dislikes     int
	UserReaction string // "like", "dislike" ou ""
	CreatedAt    time.Time
	CreatedAtStr string // Pour l'affichage formaté
}

type Category struct {
	ID   int
	Name string
}

type AllPostData struct {
	Username   string
	Posts      []PostView
	Categories []Category
}

func ServeAllPostPage(w http.ResponseWriter, r *http.Request) {
	data := AllPostData{
		Posts:      []PostView{},
		Categories: []Category{},
	}

	if cookie, err := r.Cookie("username"); err == nil {
		data.Username = cookie.Value
	}

	if db == nil {
		fmt.Println("ERREUR CRITIQUE: La connexion à la base de données est nil")
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	// Récupération des catégories
	rowsCat, err := db.Query("SELECT id, name FROM categories")
	if err == nil {
		defer rowsCat.Close()
		for rowsCat.Next() {
			var c Category
			if err := rowsCat.Scan(&c.ID, &c.Name); err == nil {
				data.Categories = append(data.Categories, c)
			}
		}
	}

	// Gestion des filtres
	categoryID := r.URL.Query().Get("category_id")
	filter := r.URL.Query().Get("filter")

	query := `
		SELECT p.id, p.title, p.content, u.username, p.created_at,
			COALESCE((
				SELECT COUNT(*) FROM reactions r 
				WHERE r.post_id = p.id AND r.type = 'like'
			), 0) as likes,
			COALESCE((
				SELECT COUNT(*) FROM reactions r 
				WHERE r.post_id = p.id AND r.type = 'dislike'
			), 0) as dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
	`
	args := []interface{}{}
	where := ""

	if categoryID != "" {
		where += "p.category_id = ?"
		args = append(args, categoryID)
	}
	if filter == "mine" && data.Username != "" {
		if where != "" {
			where += " AND "
		}
		where += "u.username = ?"
		args = append(args, data.Username)
	}
	if filter == "liked" && data.Username != "" {
		if where != "" {
			where += " AND "
		}
		where += `EXISTS (
		SELECT 1 FROM reactions r
		WHERE r.post_id = p.id AND r.type = 'like' AND r.user_id = (SELECT id FROM users WHERE username = ?)
	)`
		args = append(args, data.Username)
		// Le ORDER BY doit être ajouté après le WHERE, pas ici
	}
	if where != "" {
		query += " WHERE " + where
	}
	if filter == "liked" && data.Username != "" {
		query += " ORDER BY likes DESC, p.created_at DESC"
	} else {
		query += " ORDER BY p.created_at DESC"
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des posts: %v\nQuery: %s\n", err, query)
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post PostView
		var createdAt []byte

		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &createdAt, &post.Likes, &post.Dislikes)
		if err != nil {
			fmt.Println("Erreur lors du scan d'un post:", err)
			continue
		}

		createdAtStr := string(createdAt)
		createdAtTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			fmt.Println("Erreur lors de la conversion du timestamp:", err)
			post.CreatedAt = time.Now()
		} else {
			post.CreatedAt = createdAtTime
		}
		post.CreatedAtStr = post.CreatedAt.Format("02/01/2006 15:04")

		post.UserReaction = ""
		if data.Username != "" {
			var reactionType sql.NullString
			err := db.QueryRow(`
				SELECT type FROM reactions 
				WHERE post_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)
			`, post.ID, data.Username).Scan(&reactionType)

			if err != nil && err != sql.ErrNoRows {
				fmt.Println("Erreur lors de la récupération de la réaction:", err)
			} else if reactionType.Valid {
				post.UserReaction = reactionType.String
			}
		}

		data.Posts = append(data.Posts, post)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("Erreur lors du parcours des résultats:", err)
	}

	// Gérer les actions like/dislike si l'utilisateur est connecté
	if r.Method == "POST" && data.Username != "" {
		err := handleReaction(w, r, data.Username)
		if err != nil {
			fmt.Println("Erreur lors du traitement de la réaction:", err)
		} else {
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}
	}

	tmplPath := filepath.Join("templates", "AllPost.gohtml")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		fmt.Printf("Erreur lors du chargement du template (%s): %v\n", tmplPath, err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println("Erreur lors de l'exécution du template:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func handleReaction(w http.ResponseWriter, r *http.Request, username string) error {
	postID := r.FormValue("post_id")
	action := r.FormValue("action")

	if postID == "" || (action != "like" && action != "dislike" && action != "remove") {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("erreur début transaction: %v", err)
	}
	defer tx.Rollback()

	var existingReaction sql.NullString
	err = tx.QueryRow(`
		SELECT type FROM reactions 
		WHERE post_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)
	`, postID, username).Scan(&existingReaction)

	if action == "remove" {
		_, err = tx.Exec(`
			DELETE FROM reactions 
			WHERE post_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)
		`, postID, username)
	} else if err == nil && existingReaction.Valid {
		if existingReaction.String == action {
			return nil
		}
		_, err = tx.Exec(`
			UPDATE reactions SET type = ?
			WHERE post_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)
		`, action, postID, username)
	} else if err == sql.ErrNoRows {
		_, err = tx.Exec(`
			INSERT INTO reactions (post_id, user_id, type)
			VALUES (?, (SELECT id FROM users WHERE username = ?), ?)
		`, postID, username, action)
	}

	if err != nil {
		return fmt.Errorf("erreur exécution requête: %v", err)
	}

	return tx.Commit()
}
