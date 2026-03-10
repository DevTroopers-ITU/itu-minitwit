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