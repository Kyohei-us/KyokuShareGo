package models

import "gorm.io/gorm"

type Kyoku struct {
	gorm.Model
	Title   string   `gorm:"" json:"title"`
	Artists []Artist `gorm:"many2many:artist_kyokus;" json:"artists"`
}

type KyokuCreateJSONRequest struct {
	ArtistName string `json:"artist_name"`
	KyokuTitle string `json:"kyoku_title"`
}

type Artist struct {
	gorm.Model
	Name   string  `gorm:"" json:"name"`
	Kyokus []Kyoku `gorm:"many2many:artist_kyokus;" json:"kyokus"`
}

type ArtistCreateJSONRequest struct {
	ArtistName string `json:"artist_name"`
}

type Comment struct {
	gorm.Model
	Body    string
	KyokuID int
	Kyoku   Kyoku
	UserID  int
	User    User
}

type CommentCreateJSONRequest struct {
	Body    string `json:"comment_body"`
	KyokuID int    `json:"kyoku_id"`
	UserID  int    `json:"user_id"`
}

type CommentJoinsUser struct {
	Comment
	User
}

type User struct {
	gorm.Model
	Email          string `gorm:"unique;not null"`
	HashedPassword string
}

type UserAuthJSONRequest struct {
	Email    string `json:"user_email"`
	Password string `json:"user_password"`
}
