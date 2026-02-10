package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// GET / — personal timeline (redirect to /public if not logged in)
func timelineHandler(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	if user == nil {
		http.Redirect(w, r, "/public", http.StatusFound)
		return
	}

	messages := queryMessages(`
		SELECT message.text, message.pub_date, user.username, user.email
		FROM message, user
		WHERE message.flagged = 0 AND message.author_id = user.user_id AND (
			user.user_id = ? OR
			user.user_id IN (SELECT whom_id FROM follower WHERE who_id = ?))
		ORDER BY message.pub_date DESC LIMIT ?`,
		user.UserID, user.UserID, PER_PAGE)

	renderTemplate(w, r, "timeline.html", map[string]interface{}{
		"Messages":    messages,
		"CurrentUser": user,
		"IsTimeline":  true,
	})
}

// GET /public — public timeline
func publicTimelineHandler(w http.ResponseWriter, r *http.Request) {
	messages := queryMessages(`
		SELECT message.text, message.pub_date, user.username, user.email
		FROM message, user
		WHERE message.flagged = 0 AND message.author_id = user.user_id
		ORDER BY message.pub_date DESC LIMIT ?`, PER_PAGE)

	renderTemplate(w, r, "timeline.html", map[string]interface{}{
		"Messages": messages,
		"IsPublic": true,
	})
}

// GET /{username} — user timeline
func userTimelineHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	var profileUser User
	err := db.QueryRow("SELECT user_id, username, email, pw_hash FROM user WHERE username = ?", username).
		Scan(&profileUser.UserID, &profileUser.Username, &profileUser.Email, &profileUser.PwHash)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	followed := false
	currentUser := getCurrentUser(r)
	if currentUser != nil {
		var exists int
		err := db.QueryRow("SELECT 1 FROM follower WHERE who_id = ? AND whom_id = ?",
			currentUser.UserID, profileUser.UserID).Scan(&exists)
		followed = err == nil
	}

	messages := queryMessages(`
		SELECT message.text, message.pub_date, user.username, user.email
		FROM message, user
		WHERE user.user_id = message.author_id AND user.user_id = ?
		ORDER BY message.pub_date DESC LIMIT ?`,
		profileUser.UserID, PER_PAGE)

	renderTemplate(w, r, "timeline.html", map[string]interface{}{
		"Messages":    messages,
		"IsUser":      true,
		"ProfileUser": &profileUser,
		"Followed":    followed,
	})
}

// GET /{username}/follow
func followHandler(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	whomID := getUserID(username)
	if whomID == -1 {
		http.NotFound(w, r)
		return
	}

	db.Exec("INSERT INTO follower (who_id, whom_id) VALUES (?, ?)", user.UserID, whomID)
	addFlash(w, r, fmt.Sprintf("You are now following \"%s\"", username))
	http.Redirect(w, r, "/"+username, http.StatusFound)
}

// GET /{username}/unfollow
func unfollowHandler(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	whomID := getUserID(username)
	if whomID == -1 {
		http.NotFound(w, r)
		return
	}

	db.Exec("DELETE FROM follower WHERE who_id = ? AND whom_id = ?", user.UserID, whomID)
	addFlash(w, r, fmt.Sprintf("You are no longer following \"%s\"", username))
	http.Redirect(w, r, "/"+username, http.StatusFound)
}

// POST /add_message
func addMessageHandler(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	text := r.FormValue("text")
	if text != "" {
		db.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)",
			user.UserID, text, time.Now().Unix())
		addFlash(w, r, "Your message was recorded")
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// GET + POST /login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	if user != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	errorMsg := ""
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var u User
		err := db.QueryRow("SELECT user_id, username, email, pw_hash FROM user WHERE username = ?", username).
			Scan(&u.UserID, &u.Username, &u.Email, &u.PwHash)

		if err != nil {
			errorMsg = "Invalid username"
		} else if !checkPassword(u.PwHash, password) {
			errorMsg = "Invalid password"
		} else {
			session, _ := store.Get(r, "session")
			session.Values["user_id"] = u.UserID
			session.Save(r, w)
			addFlash(w, r, "You were logged in")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	renderTemplate(w, r, "login.html", map[string]interface{}{
		"Error": errorMsg,
	})
}

// GET + POST /register
func registerHandler(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	if user != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	errorMsg := ""
	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")

		if username == "" {
			errorMsg = "You have to enter a username"
		} else if email == "" || !strings.Contains(email, "@") {
			errorMsg = "You have to enter a valid email address"
		} else if password == "" {
			errorMsg = "You have to enter a password"
		} else if password != password2 {
			errorMsg = "The two passwords do not match"
		} else if getUserID(username) != -1 {
			errorMsg = "The username is already taken"
		} else {
			db.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)",
				username, email, hashPassword(password))
			addFlash(w, r, "You were successfully registered and can login now")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
	}

	renderTemplate(w, r, "register.html", map[string]interface{}{
		"Error": errorMsg,
	})
}

// GET /logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	addFlash(w, r, "You were logged out")
	http.Redirect(w, r, "/public", http.StatusFound)
}
