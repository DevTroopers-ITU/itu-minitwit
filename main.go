package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gorm.io/gorm"
	"github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Configuration
const (
	DATABASE   = "/tmp/minitwit.db"
	PER_PAGE   = 30
	SECRET_KEY = "development key"
)

// Globals
var (
	db           *gorm.DB
	store        *DBStore
	sessionStore *sessions.CookieStore
	// Prometheus counter
    httpResponsesTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "minitwit_http_responses_total",
        Help: "Total number of HTTP responses",
    })
)

// metricsMiddleware wraps the router and increments the Prometheus
// httpResponsesTotal counter on every incoming HTTP request.
// It calls next.ServeHTTP to pass the request to the actual route handler,
// and then increments the counter after the handler has finished.
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        next.ServeHTTP(w, r)
        httpResponsesTotal.Inc()
    })
}

// Router setup
func setupRouter() *mux.Router {
	// Register metric
	prometheus.MustRegister(httpResponsesTotal)
	r := mux.NewRouter()
	// Add metrics endpoint
    r.Handle("/metrics", promhttp.Handler())

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

func main() {
	initDB()
	store = NewDBStore(db)
	sessionStore = newStore()

	r := setupRouter()

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", metricsMiddleware(r)))
}