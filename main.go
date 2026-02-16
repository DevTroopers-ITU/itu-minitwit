package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

// Configuration
const (
	DATABASE   = "/tmp/minitwit.db"
	PER_PAGE   = 30
	SECRET_KEY = "development key"
)

// Globals
var (
	db    *sql.DB
	store *sessions.CookieStore
)

// Router setup
func setupRouter() *mux.Router {
	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Swagger docs
	r.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger.html")
	}).Methods("GET")
	r.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger.json")
	}).Methods("GET")

	// Sim API (before /{username} catch-all)
	r.HandleFunc("/latest", getLatest).Methods("GET")
	r.HandleFunc("/register", simRegister).Methods("POST").Headers("Content-Type", "application/json")
	r.HandleFunc("/msgs", simMessages).Methods("GET")
	r.HandleFunc("/msgs/{username}", simMessagesPerUser).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", simFollow).Methods("GET", "POST")

	// Web UI routes
	r.HandleFunc("/public", publicTimelineHandler).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", registerHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")
	r.HandleFunc("/add_message", addMessageHandler).Methods("POST")

	// User routes (catch-all — must be last)
	r.HandleFunc("/{username}/follow", followHandler).Methods("GET")
	r.HandleFunc("/{username}/unfollow", unfollowHandler).Methods("GET")
	r.HandleFunc("/{username}", userTimelineHandler).Methods("GET")

	// Root
	r.HandleFunc("/", timelineHandler).Methods("GET")
	return r
}

func main() {
	initDB()
	store = newStore()

	r := setupRouter()

	log.Println("Listening on http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", r))
}
