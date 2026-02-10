package main

// User represents a registered user.
type User struct {
	UserID   int
	Username string
	Email    string
	PwHash   string
}

// Message represents a tweet/message joined with user info.
type Message struct {
	Username string
	Email    string
	Text     string
	PubDate  int64
}
