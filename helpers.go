package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// --- Session helpers ---

func newStore() *sessions.CookieStore {
	s := sessions.NewCookieStore([]byte(SECRET_KEY))
	s.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
	}
	return s
}

func getCurrentUser(r *http.Request) *User {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"]
	if !ok {
		return nil
	}
	return getUserByID(userID.(int))
}

func addFlash(w http.ResponseWriter, r *http.Request, message string) {
	session, _ := store.Get(r, "session")
	session.AddFlash(message)
	session.Save(r, w)
}

func getFlashes(w http.ResponseWriter, r *http.Request) []interface{} {
	session, _ := store.Get(r, "session")
	flashes := session.Flashes()
	session.Save(r, w)
	return flashes
}

// --- Password helpers ---

func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// --- Template helpers ---

func gravatar(email string) string {
	h := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=identicon&s=48", h)
}

func datetimeformat(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02 @ 15:04")
}

func renderTemplate(w http.ResponseWriter, r *http.Request, templateFile string, data map[string]interface{}) {
	funcMap := template.FuncMap{
		"gravatar":       gravatar,
		"datetimeformat": datetimeformat,
	}

	tmpl := template.Must(template.New("layout.html").
		Funcs(funcMap).
		ParseFiles("templates/layout.html", "templates/"+templateFile))

	if _, ok := data["CurrentUser"]; !ok {
		data["CurrentUser"] = getCurrentUser(r)
	}
	if _, ok := data["Flashes"]; !ok {
		data["Flashes"] = getFlashes(w, r)
	}

	tmpl.ExecuteTemplate(w, "layout.html", data)
}
