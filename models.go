package main

// User represents a registered user.
type User struct {
	UserID   int `gorm:"primaryKey"`
	Username string
	Email    string
	PwHash   string
}

func (User) TableName() string {
	return "user"
}

// Message represents a tweet/message joined with user info.
type Message struct {
	MessageID int `gorm:"primaryKey"`
	AuthorID  int
	Author    User `gorm:"foreignKey:AuthorID;references:UserID"`
	Text      string
	PubDate   int64
	Flagged   int
}

func (Message) TableName() string {
	return "message"
}

// Follower represents a follow relation.
type Follower struct {
	WhoID  int
	WhomID int
}

func (Follower) TableName() string {
	return "follower"
}
