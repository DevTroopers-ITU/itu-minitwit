package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	Username string
	Email    string
	Text     string
	PubDate  int64
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "/tmp/minitwit.db") // adjust path
	if err != nil {
		log.Fatal(err)
	}
}

func gravatar(email string) string {
	h := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=identicon&s=48", h)
}

func datetimeformat(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02 15:04")
}

func publicTimelineHandler(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"gravatar":       gravatar,
		"datetimeformat": datetimeformat,
	}

	rows, err := db.Query(`
        SELECT message.text, message.pub_date, user.username, user.email
        FROM message
        JOIN user ON message.author_id = user.user_id
        WHERE message.flagged = 0
        ORDER BY message.pub_date DESC
        LIMIT 30
    `)
	if err != nil {
		http.Error(w, "DB error", 500)
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

	tmpl := template.Must(template.New("layout.html").
		Funcs(funcMap).
		ParseFiles("templates/layout.html", "templates/timeline.html"))

	data := map[string]interface{}{
		"Messages":    messages,
		"CurrentUser": map[string]interface{}{"Username": "bob"}, // replace with actual login
		"IsPublic":    true,
		"IsTimeline":  true,
	}

	tmpl.ExecuteTemplate(w, "layout.html", data)
}

func main() {
	initDB() // connect to SQLite

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/public", publicTimelineHandler)

	log.Println("Listening on http://localhost:5000/public")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
