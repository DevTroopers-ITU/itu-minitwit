package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

// Configuration
const (
	PER_PAGE = 30
)

// Globals
var (
	db           *gorm.DB
	store        *DBStore
	sessionStore *sessions.CookieStore
	// Prometheus metrics
	httpResponsesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "minitwit_http_responses_total",
		Help: "Total number of HTTP responses",
	}, []string{"method", "route", "status"})

	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "minitwit_http_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
	SECRET_KEY = getSecretKey()
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start).Seconds()

		// Get the matched route template instead of the actual path
		route := r.URL.Path
		if routeMatch := mux.CurrentRoute(r); routeMatch != nil {
			if template, err := routeMatch.GetPathTemplate(); err == nil {
				route = template
			}
		}

		httpResponsesTotal.WithLabelValues(r.Method, route, fmt.Sprintf("%d", wrapped.status)).Inc()
		httpDuration.WithLabelValues(r.Method, route).Observe(duration)
	})
}

// Router setup
func setupRouter() *mux.Router {

	r := mux.NewRouter()
	r.Use(metricsMiddleware)
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

func getSecretOrEnv(name string) string {
	path := "/run/secrets/" + strings.ToLower(name)
	if data, err := os.ReadFile(path); err == nil {
		return strings.TrimSpace(string(data))
	}
	return os.Getenv(name)
}

func getSecretKey() string {
	if err := godotenv.Load(); err == nil {
		if key := getSecretOrEnv("SECRET_KEY"); key != "" {
			return key
		}
	}
	return "dev-fallback-key-change-in-production"
}

func main() {
	initDB()
	prometheus.MustRegister(httpResponsesTotal)
	prometheus.MustRegister(httpDuration)
	store = NewDBStore(db)
	sessionStore = newStore()

	r := setupRouter()

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
