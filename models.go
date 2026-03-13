package main

// User represents a registered user.
type User struct {
	UserID   int `gorm:"primaryKey"`
	Username string
	Email    string
	PwHash   string
}

func (User) TableName() string {
    return "users"
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


// Follower represents a follow relation.
type Follower struct {
	WhoID  int
	WhomID int
}

func (Message) TableName() string {
    return "messages"
}

func (Follower) TableName() string {
    return "followers"
}
