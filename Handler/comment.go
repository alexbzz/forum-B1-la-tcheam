package Handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Comment struct {
	ID        int    `json:"id"`
	PostID    int    `json:"post_id"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// Handler pour ajouter un commentaire (POST)
func AddComment(w http.ResponseWriter, r *http.Request) {
	// Vérifie si l'utilisateur est connecté via cookie
	var username string
	if cookie, err := r.Cookie("username"); err == nil {
		username = cookie.Value
	} else {
		fmt.Println("DEBUG: Utilisateur non connecté")
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérification de la connexion DB
	if db == nil {
		fmt.Println("ERREUR CRITIQUE: La connexion à la base de données est nil")
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	// Test de la connexion DB
	if err := db.Ping(); err != nil {
		fmt.Printf("ERREUR: Connexion DB fermée: %v\n", err)
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	// Récupération des données du formulaire
	postIDStr := r.FormValue("post_id")
	content := r.FormValue("content")

	fmt.Printf("DEBUG: Données reçues - post_id: %s, content: %s, username: %s\n", postIDStr, content, username)

	// Validation des données
	if postIDStr == "" || content == "" {
		fmt.Println("DEBUG: Données manquantes")
		http.Error(w, "Données manquantes", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		fmt.Printf("DEBUG: Erreur conversion post_id: %v\n", err)
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}

	// Vérifier que l'utilisateur existe et récupérer son ID
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		fmt.Printf("DEBUG: Utilisateur non trouvé: %v\n", err)
		http.Error(w, "Utilisateur non trouvé", http.StatusUnauthorized)
		return
	}

	fmt.Printf("DEBUG: UserID trouvé: %d\n", userID)

	// Vérifier que le post existe
	var postExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", postID).Scan(&postExists)
	if err != nil || !postExists {
		fmt.Printf("DEBUG: Post non trouvé - erreur: %v, exists: %v\n", err, postExists)
		http.Error(w, "Post non trouvé", http.StatusNotFound)
		return
	}

	fmt.Printf("DEBUG: Post %d existe\n", postID)

	// Transaction pour l'insertion
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("Erreur début transaction: %v\n", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}

	// Gestion simplifiée du rollback
	defer func() {
		if err != nil {
			fmt.Printf("Rollback transaction à cause de l'erreur: %v\n", err)
			tx.Rollback()
		}
	}()

	// Insertion du commentaire directement
	result, err := tx.Exec(
		"INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)",
		postID, userID, content,
	)
	if err != nil {
		fmt.Printf("Erreur lors de l'insertion du commentaire: %v\n", err)
		http.Error(w, "Erreur lors de l'ajout du commentaire", http.StatusInternalServerError)
		return
	}

	// Vérifier que l'insertion a bien eu lieu
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Printf("Erreur lors de la vérification des lignes affectées: %v\n", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}

	fmt.Printf("DEBUG: Lignes affectées: %d\n", rowsAffected)
	if rowsAffected == 0 {
		fmt.Println("ATTENTION: Aucune ligne affectée par l'insertion")
		http.Error(w, "Échec de l'insertion", http.StatusInternalServerError)
		return
	}

	// Récupérer l'ID du commentaire inséré
	commentID, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("Erreur récupération ID: %v\n", err)
	} else {
		fmt.Printf("DEBUG: Commentaire inséré avec ID: %d\n", commentID)
	}

	// Commit de la transaction
	if err = tx.Commit(); err != nil {
		fmt.Printf("Erreur lors du commit: %v\n", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}

	fmt.Printf("DEBUG: Commentaire ajouté avec succès pour le post %d (transaction commitée)\n", postID)

	// Redirection vers la page du post
	http.Redirect(w, r, fmt.Sprintf("/post?id=%d", postID), http.StatusSeeOther)
}

// Handler pour charger les commentaires d'un post (GET, JSON)
func GetComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérification de la connexion DB
	if db == nil {
		fmt.Println("ERREUR CRITIQUE: La connexion à la base de données est nil")
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	// Test de la connexion DB
	if err := db.Ping(); err != nil {
		fmt.Printf("ERREUR: Connexion DB fermée: %v\n", err)
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		http.Error(w, "ID de post manquant", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}

	// Récupération des commentaires
	query := `
		SELECT c.id, c.post_id, u.username, c.content, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC
	`

	rows, err := db.Query(query, postID)
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des commentaires: %v\n", err)
		http.Error(w, "Erreur lors de la récupération des commentaires", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var createdAt time.Time

		err := rows.Scan(&comment.ID, &comment.PostID, &comment.Author, &comment.Content, &createdAt)
		if err != nil {
			fmt.Printf("Erreur lors du scan d'un commentaire: %v\n", err)
			continue
		}

		// Formatage de la date
		comment.CreatedAt = createdAt.Format("02/01/2006 15:04")
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		fmt.Printf("Erreur lors du parcours des résultats: %v\n", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}

	fmt.Printf("DEBUG: %d commentaires trouvés pour le post %d\n", len(comments), postID)

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(comments); err != nil {
		fmt.Printf("Erreur lors de l'encodage JSON: %v\n", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
	}
}
