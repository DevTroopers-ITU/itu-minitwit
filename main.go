package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbPath  = "/tmp/minitwit.db"
	perPage = 30
)

type Message struct {
	Text     string
	PubDate  int64
	Username string
	Email    string
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/public", publicTimeline)

	log.Println("Listening on http://localhost:5000/public")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func publicTimeline(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT message.text, message.pub_date, user.username, user.email
		FROM message
		JOIN user ON message.author_id = user.user_id
		WHERE message.flagged = 0
		ORDER BY message.pub_date DESC
		LIMIT ?`, perPage)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.Text, &m.PubDate, &m.Username, &m.Email); err != nil {
			http.Error(w, "Scan error", 500)
			return
		}
		messages = append(messages, m)
	}

	tmpl := template.Must(template.New("timeline").Parse(`
		<!doctype html>
		<html>
		<head><title>Public Timeline</title></head>
		<body>
			<h1>Public Timeline</h1>
			{{ range . }}
				<div>
					<strong>{{ .Username }}</strong><br>
					{{ .Text }}<br>
					<small>{{ .PubDate }}</small>
				</div>
				<hr>
			{{ else }}
				<p>No messages yet.</p>
			{{ end }}
		</body>
		</html>
	`))

	tmpl.Execute(w, messages)
}
