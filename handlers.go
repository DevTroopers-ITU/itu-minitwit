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

	messages, err := store.PersonalTimeline(user.UserID, PER_PAGE)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, r, "timeline.html", map[string]interface{}{
		"Messages":    messages,
		"CurrentUser": user,
		"IsTimeline":  true,
	})
}

// GET /public — public timeline
func publicTimelineHandler(w http.ResponseWriter, r *http.Request) {
	messages, err := store.PublicTimeline(PER_PAGE)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, r, "timeline.html", map[string]interface{}{
		"Messages": messages,
		"IsPublic": true,
	})
}

func registerDispatcher(w http.ResponseWriter, r *http.Request) {
    ct := r.Header.Get("Content-Type")
    if strings.HasPrefix(ct, "application/json") {
        simRegister(w, r)
    } else {
        registerHandler(w, r)
    }
}

// GET /{username} — user timeline
func userTimelineHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	profileUser, err := store.GetUserByUsername(username)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	followed := false
	currentUser := getCurrentUser(r)
	if currentUser != nil {
		followed = store.IsFollowing(currentUser.UserID, profileUser.UserID)
	}

	messages, err := store.UserTimeline(profileUser.UserID, PER_PAGE)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, r, "timeline.html", map[string]interface{}{
		"Messages":    messages,
		"IsUser":      true,
		"ProfileUser": profileUser,
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
	whomID := store.GetUserID(username)
	if whomID == -1 {
		http.NotFound(w, r)
		return
	}

	store.Follow(user.UserID, whomID)
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
	whomID := store.GetUserID(username)
	if whomID == -1 {
		http.NotFound(w, r)
		return
	}

	store.Unfollow(user.UserID, whomID)
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
		store.AddMessage(user.UserID, text, time.Now().Unix())
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

		u, err := store.GetUserByUsername(username)
		if err != nil {
			errorMsg = "Invalid username"
		} else if !checkPassword(u.PwHash, password) {
			errorMsg = "Invalid password"
		} else {
			session, _ := sessionStore.Get(r, "session")
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
		} else if store.GetUserID(username) != -1 {
			errorMsg = "The username is already taken"
		} else {
			store.CreateUser(username, email, hashPassword(password))
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
	session, _ := sessionStore.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	addFlash(w, r, "You were logged out")
	http.Redirect(w, r, "/public", http.StatusFound)
}