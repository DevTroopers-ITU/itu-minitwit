package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initDB() {
	var err error

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		log.Println("Connecting to PostgreSQL")
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else {
		log.Println("Connecting to SQLite:", DATABASE)
		db, err = gorm.Open(sqlite.Open(DATABASE), &gorm.Config{})
	}
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&User{}, &Message{}, &Follower{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}