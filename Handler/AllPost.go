package Handler

import (
	"fmt"
	"net/http"
	"time"
)

type AllPostsData struct {
	Username string
	Posts    []PostWithAuthor
}

type PostWithAuthor struct {
	ID        int
	Title     string
	Content   string
	Author    string
	CreatedAt time.Time
}

func ListAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	username := ""
	cookie, err := r.Cookie("username")
	if err == nil {
		username = cookie.Value
	}

	rows, err := db.Query(`
        SELECT p.id, p.title, p.content, p.created_at, u.username
        FROM posts p
        JOIN users u ON p.user_id = u.id
        ORDER BY p.created_at DESC
    `)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []PostWithAuthor
	for rows.Next() {
		var post PostWithAuthor
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.Author)
		if err != nil {
			fmt.Println("Erreur lors du scan d'un post:", err)
			continue
		}
		posts = append(posts, post)
	}

	data := AllPostsData{
		Username: username,
		Posts:    posts,
	}

	tmpl, err := template.ParseFiles("./forum-B1-la-tcheam/templates/all-posts.gohtml")
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}
