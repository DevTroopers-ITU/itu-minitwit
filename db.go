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

	createIndexes()
}

func createIndexes() {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_messages_flagged_pubdate ON messages (flagged, pub_date DESC)",
		"CREATE INDEX IF NOT EXISTS idx_messages_author_pubdate ON messages (author_id, pub_date DESC)",
		"CREATE INDEX IF NOT EXISTS idx_follower_who_id ON follower (who_id)",
	}
	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Printf("Warning: failed to create index: %v", err)
		}
	}
}
