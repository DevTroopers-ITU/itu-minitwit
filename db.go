package main

import (
	"log"
	"time"

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

    // Add this block:
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatal(err)
    }
    sqlDB.SetMaxOpenConns(25)
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)

    err = db.AutoMigrate(&User{}, &Message{}, &Follower{})
    if err != nil {
        log.Fatal("Failed to migrate database:", err)
    }
}