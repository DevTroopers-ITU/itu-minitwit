package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDB() {
	var err error
	dsn := getSecretOrEnv("DATABASE_URL")

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&User{}, &Message{}, &Follower{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
