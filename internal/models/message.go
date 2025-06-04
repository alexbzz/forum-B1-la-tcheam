package models

import (
	"database/sql"
	"time"
)

type PrivateMessage struct {
	ID      int
	Sender  string
	Content string
	IsRead  bool
	SentAt  time.Time
}

func GetInboxMessages(db *sql.DB, userID int) ([]PrivateMessage, error) {
	rows, err := db.Query(`
		SELECT pm.id, u.username AS sender, pm.content, pm.is_read, pm.sent_at
		FROM private_messages pm
		JOIN users u ON pm.sender_id = u.id
		WHERE pm.receiver_id = ?
		ORDER BY pm.sent_at DESC //exemple de db a changer avec la vrai
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []PrivateMessage
	for rows.Next() {
		var msg PrivateMessage
		err := rows.Scan(&msg.ID, &msg.Sender, &msg.Content, &msg.IsRead, &msg.SentAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
