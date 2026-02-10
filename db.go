package main

import (
	"database/sql"
	"log"
)

func openDB(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", path)
}

func initDB() {
	var err error
	db, err = openDB(DATABASE)
	if err != nil {
		log.Fatal(err)
	}
}

func getUserByID(userID int) *User {
	var u User
	err := db.QueryRow("SELECT user_id, username, email, pw_hash FROM user WHERE user_id = ?", userID).
		Scan(&u.UserID, &u.Username, &u.Email, &u.PwHash)
	if err != nil {
		return nil
	}
	return &u
}

func getUserID(username string) int {
	var id int
	err := db.QueryRow("SELECT user_id FROM user WHERE username = ?", username).Scan(&id)
	if err != nil {
		return -1
	}
	return id
}

func queryMessages(query string, args ...interface{}) []Message {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.Text, &m.PubDate, &m.Username, &m.Email); err != nil {
			continue
		}
		messages = append(messages, m)
	}
	return messages
}
