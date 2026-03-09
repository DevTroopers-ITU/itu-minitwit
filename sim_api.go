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
	if store.GetUserID(data["username"]) != -1 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "error_msg": "The username is already taken"})
		return
	}

	store.CreateUser(data["username"], data["email"], hashPassword(data["pwd"]))
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

	messages, err := store.PublicTimeline(noMsgs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	filtered := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		filtered = append(filtered, map[string]interface{}{
			"content":  msg.Text,
			"pub_date": msg.PubDate,
			"user":     msg.Username,
		})
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
	userID := store.GetUserID(username)
	if userID == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		noMsgs, _ := strconv.Atoi(r.URL.Query().Get("no"))
		if noMsgs == 0 {
			noMsgs = 100
		}

		messages, err := store.UserTimeline(userID, noMsgs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		filtered := make([]map[string]interface{}, 0, len(messages))
		for _, msg := range messages {
			filtered = append(filtered, map[string]interface{}{
				"content":  msg.Text,
				"pub_date": msg.PubDate,
				"user":     msg.Username,
			})
		}
		json.NewEncoder(w).Encode(filtered)

	case "POST":
		var data map[string]string
		json.NewDecoder(r.Body).Decode(&data)
		store.AddMessage(userID, data["content"], time.Now().Unix())
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
	userID := store.GetUserID(username)
	if userID == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.Method {
	case "POST":
		var data map[string]string
		json.NewDecoder(r.Body).Decode(&data)

		if followUser, ok := data["follow"]; ok {
			followID := store.GetUserID(followUser)
			if followID == -1 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			store.Follow(userID, followID)
			w.WriteHeader(http.StatusNoContent)

		} else if unfollowUser, ok := data["unfollow"]; ok {
			unfollowID := store.GetUserID(unfollowUser)
			if unfollowID == -1 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			store.Unfollow(userID, unfollowID)
			w.WriteHeader(http.StatusNoContent)
		}

	case "GET":
		noFollowers, _ := strconv.Atoi(r.URL.Query().Get("no"))
		if noFollowers == 0 {
			noFollowers = 100
		}

		follows, err := store.FollowingUsernames(userID, noFollowers)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"follows": follows})
	}
}