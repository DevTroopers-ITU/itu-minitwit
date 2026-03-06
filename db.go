package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initDB() {
	var err error

	db, err = gorm.Open(sqlite.Open(DATABASE), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&User{}, &Message{}, &Follower{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}

func getUserByID(userID int) *User {
	var u User
	result := db.First(&u, "user_id = ?", userID)
	if result.Error != nil {
		return nil
	}
	return &u
}

func getUserID(username string) int {
	var u User
	result := db.First(&u, "username = ?", username)
	if result.Error != nil {
		return -1
	}
	return u.UserID
}

func queryMessages(query string, args ...interface{}) []MessageView {
	var messages []MessageView
	result := db.Raw(query, args...).Scan(&messages)
	if result.Error != nil {
		return nil
	}
	return messages
}
