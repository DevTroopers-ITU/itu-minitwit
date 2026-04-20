package main

import "gorm.io/gorm"

type DBStore struct {
	db *gorm.DB
}

func NewDBStore(db *gorm.DB) *DBStore {
	return &DBStore{db: db}
}

// Users

func (s *DBStore) GetUserByID(userID int) (*User, error) {
	var u User
	if err := s.db.First(&u, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *DBStore) GetUserByUsername(username string) (*User, error) {
	var u User
	if err := s.db.First(&u, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *DBStore) GetUserID(username string) int {
	u, err := s.GetUserByUsername(username)
	if err != nil {
		return -1
	}
	return u.UserID
}

func (s *DBStore) CreateUser(username, email, pwHash string) error {
	u := User{Username: username, Email: email, PwHash: pwHash}
	return s.db.Create(&u).Error
}

// Follows

func (s *DBStore) IsFollowing(whoID, whomID int) bool {
	var f Follower
	err := s.db.First(&f, "who_id = ? AND whom_id = ?", whoID, whomID).Error
	return err == nil
}

func (s *DBStore) Follow(whoID, whomID int) error {
	f := Follower{WhoID: whoID, WhomID: whomID}
	return s.db.FirstOrCreate(&f, f).Error
}

func (s *DBStore) Unfollow(whoID, whomID int) error {
	return s.db.Where("who_id = ? AND whom_id = ?", whoID, whomID).Delete(&Follower{}).Error
}

func (s *DBStore) FollowingUsernames(whoID, limit int) ([]string, error) {
	var fs []Follower
	if err := s.db.Where("who_id = ?", whoID).Limit(limit).Find(&fs).Error; err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(fs))
	for _, f := range fs {
		ids = append(ids, f.WhomID)
	}
	if len(ids) == 0 {
		return []string{}, nil
	}

	var users []User
	if err := s.db.Select("username").Where("user_id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}

	out := make([]string, 0, len(users))
	for _, u := range users {
		out = append(out, u.Username)
	}
	return out, nil
}

// Messages

func (s *DBStore) AddMessage(authorID int, text string, pubDate int64) error {
	m := Message{AuthorID: authorID, Text: text, PubDate: pubDate, Flagged: 0}
	return s.db.Create(&m).Error
}

// DTO til templates/sim-api (matcher jeres eksisterende brug)
type MessageView struct {
	Username string
	Email    string
	Text     string
	PubDate  int64
}

func toViews(msgs []Message) []MessageView {
	out := make([]MessageView, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, MessageView{
			Username: m.Author.Username,
			Email:    m.Author.Email,
			Text:     m.Text,
			PubDate:  m.PubDate,
		})
	}
	return out
}

func (s *DBStore) PublicTimeline(limit int) ([]MessageView, error) {
	var msgs []Message
	err := s.db.
		Where("flagged = ?", 0).
		Preload("Author").
		Order("pub_date desc").
		Limit(limit).
		Find(&msgs).Error
	return toViews(msgs), err
}

func (s *DBStore) UserTimeline(userID, limit int) ([]MessageView, error) {
	var msgs []Message
	err := s.db.
		Where("author_id = ?", userID).
		Preload("Author").
		Order("pub_date desc").
		Limit(limit).
		Find(&msgs).Error
	return toViews(msgs), err
}

func (s *DBStore) PersonalTimeline(userID, limit int) ([]MessageView, error) {
    var msgs []Message
    err := s.db.
        Joins("LEFT JOIN followers ON messages.author_id = followers.whom_id AND followers.who_id = ?", userID).
        Where("messages.flagged = 0").
        Where("messages.author_id = ? OR followers.who_id IS NOT NULL", userID).
        Preload("Author").
        Order("messages.pub_date desc").
        Limit(limit).
        Find(&msgs).Error
    return toViews(msgs), err
}