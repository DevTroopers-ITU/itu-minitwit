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

// SimState holds singleton state for the simulator API. The row with ID=1 is
// the only row used; other rows must never be written. Backing the `latest`
// counter in the DB (rather than a package-level var) keeps the value
// consistent across webserver replicas.
type SimState struct {
	ID     int `gorm:"primaryKey"`
	Latest int
}

func (SimState) TableName() string {
	return "sim_states"
}
