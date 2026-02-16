package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var latest int = -1

// helper functions

func getLatest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"latest": latest})
}

func updateLatest(r *http.Request) {
	if latestStr := r.URL.Query().Get("latest"); latestStr != "" {
		if latestInt, err := strconv.Atoi(latestStr); err == nil {
			latest = latestInt
		}
	}
}

func notReqFromSimulator(r *http.Request) bool {
	return r.Header.Get("Authorization") != "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh"
}

// Sim API handlers

func simRegister(w http.ResponseWriter, r *http.Request) {
	updateLatest(r)
	w.Header().Set("Content-Type", "application/json")

	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "error_msg": "Bad request"})
		return
	}

	if data["username"] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "error_msg": "You have to enter a username"})
		return
	}
	if data["email"] == "" || !strings.Contains(data["email"], "@") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "error_msg": "You have to enter a valid email address"})
		return
	}
	if data["pwd"] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "error_msg": "You have to enter a password"})
		return
	}
	if getUserID(data["username"]) != -1 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "error_msg": "The username is already taken"})
		return
	}

	hashedPassword := hashPassword(data["pwd"])
	db.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", data["username"], data["email"], hashedPassword)
	w.WriteHeader(http.StatusNoContent)
}

func simMessages(w http.ResponseWriter, r *http.Request) {
	updateLatest(r)
	w.Header().Set("Content-Type", "application/json")

	if notReqFromSimulator(r) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 403, "error_msg": "You are not authorized to use this resource!"})
		return
	}

	noMsgs, _ := strconv.Atoi(r.URL.Query().Get("no"))
	if noMsgs == 0 {
		noMsgs = 100
	}

	query := `SELECT message.text, message.pub_date, user.username, user.email
			  FROM message, user
			  WHERE message.flagged = 0 AND message.author_id = user.user_id
			  ORDER BY message.pub_date DESC LIMIT ?`
	messages := queryMessages(query, noMsgs)

	var filtered []map[string]interface{}
	for _, msg := range messages {
		filtered = append(filtered, map[string]interface{}{
			"content":  msg.Text,
			"pub_date": msg.PubDate,
			"user":     msg.Username,
		})
	}

	if filtered == nil {
		filtered = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(filtered)
}

func simMessagesPerUser(w http.ResponseWriter, r *http.Request) {
	updateLatest(r)
	w.Header().Set("Content-Type", "application/json")

	if notReqFromSimulator(r) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 403, "error_msg": "You are not authorized to use this resource!"})
		return
	}

	username := mux.Vars(r)["username"]
	userID := getUserID(username)
	if userID == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		noMsgs, _ := strconv.Atoi(r.URL.Query().Get("no"))
		if noMsgs == 0 {
			noMsgs = 100
		}

		query := `SELECT message.text, message.pub_date, user.username, user.email
				  FROM message, user
				  WHERE message.flagged = 0 AND user.user_id = message.author_id
				  AND user.user_id = ?
				  ORDER BY message.pub_date DESC LIMIT ?`
		messages := queryMessages(query, userID, noMsgs)

		var filtered []map[string]interface{}
		for _, msg := range messages {
			filtered = append(filtered, map[string]interface{}{
				"content":  msg.Text,
				"pub_date": msg.PubDate,
				"user":     msg.Username,
			})
		}

		if filtered == nil {
			filtered = []map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(filtered)

	} else if r.Method == "POST" {
		var data map[string]string
		json.NewDecoder(r.Body).Decode(&data)

		db.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)",
			userID, data["content"], time.Now().Unix())
		w.WriteHeader(http.StatusNoContent)
	}
}

func simFollow(w http.ResponseWriter, r *http.Request) {
	updateLatest(r)
	w.Header().Set("Content-Type", "application/json")

	if notReqFromSimulator(r) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 403, "error_msg": "You are not authorized to use this resource!"})
		return
	}

	username := mux.Vars(r)["username"]
	userID := getUserID(username)
	if userID == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == "POST" {
		var data map[string]string
		json.NewDecoder(r.Body).Decode(&data)

		if followUser, ok := data["follow"]; ok {
			followID := getUserID(followUser)
			if followID == -1 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			db.Exec("INSERT INTO follower (who_id, whom_id) VALUES (?, ?)", userID, followID)
			w.WriteHeader(http.StatusNoContent)

		} else if unfollowUser, ok := data["unfollow"]; ok {
			unfollowID := getUserID(unfollowUser)
			if unfollowID == -1 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			db.Exec("DELETE FROM follower WHERE who_id=? AND whom_id=?", userID, unfollowID)
			w.WriteHeader(http.StatusNoContent)
		}

	} else if r.Method == "GET" {
		noFollowers, _ := strconv.Atoi(r.URL.Query().Get("no"))
		if noFollowers == 0 {
			noFollowers = 100
		}

		rows, err := db.Query(`SELECT user.username FROM user
							   INNER JOIN follower ON follower.whom_id = user.user_id
							   WHERE follower.who_id = ?
							   LIMIT ?`, userID, noFollowers)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var follows []string
		for rows.Next() {
			var name string
			rows.Scan(&name)
			follows = append(follows, name)
		}

		if follows == nil {
			follows = []string{}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"follows": follows})
	}
}
