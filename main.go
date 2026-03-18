package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// Configuration
const (
	DATABASE = "/tmp/minitwit.db"
	PER_PAGE = 30
)

// Globals
var (
	db           *gorm.DB
	store        *DBStore
	sessionStore *sessions.CookieStore
	SECRET_KEY   = getSecretKey()
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
	r.HandleFunc("/msgs", simMessages).Methods("GET")
	r.HandleFunc("/msgs/{username}", simMessagesPerUser).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", simFollow).Methods("GET", "POST")

	// Web UI routes
	r.HandleFunc("/public", publicTimelineHandler).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", registerDispatcher).Methods("GET", "POST")
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

func getSecretKey() string {
	if err := godotenv.Load(); err == nil {
		if key := os.Getenv("SECRET_KEY"); key != "" {
			return key
		}
	}
	return "dev-fallback-key-change-in-production"
}

func main() {
	initDB()
	store = NewDBStore(db)
	sessionStore = newStore()

	r := setupRouter()

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
