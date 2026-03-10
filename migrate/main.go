package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	UserID   int `gorm:"primaryKey"`
	Username string
	Email    string
	PwHash   string
}

func (User) TableName() string { return "user" }

type Message struct {
	MessageID int `gorm:"primaryKey"`
	AuthorID  int
	Text      string
	PubDate   int64
	Flagged   int
}

func (Message) TableName() string { return "message" }

type Follower struct {
	WhoID  int
	WhomID int
}

func (Follower) TableName() string { return "follower" }

func main() {
	sqlitePath := "/tmp/minitwit.db"
	pgDSN := os.Getenv("DATABASE_URL")
	if pgDSN == "" {
		log.Fatal("DATABASE_URL is required")
	}

	sqliteDB, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to open SQLite:", err)
	}

	pgDB, err := gorm.Open(postgres.Open(pgDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to open PostgreSQL:", err)
	}

	pgDB.AutoMigrate(&User{}, &Message{}, &Follower{})

	// Migrate users
	var userCount int64
	sqliteDB.Model(&User{}).Count(&userCount)
	fmt.Printf("Migrating %d users...\n", userCount)

	var users []User
	sqliteDB.FindInBatches(&users, 1000, func(tx *gorm.DB, batch int) error {
		pgDB.Create(&users)
		fmt.Printf("  Users batch %d done\n", batch)
		return nil
	})

	// Migrate messages
	var msgCount int64
	sqliteDB.Model(&Message{}).Count(&msgCount)
	fmt.Printf("Migrating %d messages...\n", msgCount)

	var msgs []Message
	sqliteDB.FindInBatches(&msgs, 5000, func(tx *gorm.DB, batch int) error {
		pgDB.Create(&msgs)
		fmt.Printf("  Messages batch %d done\n", batch)
		return nil
	})

	// Migrate followers
	var followCount int64
	sqliteDB.Model(&Follower{}).Count(&followCount)
	fmt.Printf("Migrating %d followers...\n", followCount)

	var follows []Follower
	sqliteDB.FindInBatches(&follows, 1000, func(tx *gorm.DB, batch int) error {
		pgDB.Create(&follows)
		fmt.Printf("  Followers batch %d done\n", batch)
		return nil
	})

	// Fix PostgreSQL sequences so new IDs continue after migrated data
	pgDB.Exec(`SELECT setval(pg_get_serial_sequence('"user"', 'user_id'), (SELECT COALESCE(MAX(user_id), 1) FROM "user"))`)
	pgDB.Exec(`SELECT setval(pg_get_serial_sequence('message', 'message_id'), (SELECT COALESCE(MAX(message_id), 1) FROM message))`)

	fmt.Println("Migration complete!")
}
