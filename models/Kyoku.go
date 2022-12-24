package models

import "gorm.io/gorm"

type Kyoku struct {
	gorm.Model
	Title string `gorm:"" json:"title"`
}

type KyokuCreateJSONRequest struct {
	ArtistName string `json:"artist_name"`
	KyokuTitle string `json:"kyoku_title"`
}

type ArtistCreateJSONRequest struct {
	ArtistName string `json:"artist_name"`
}

type Artist struct {
	gorm.Model
	Name   string  `gorm:"" json:"name"`
	Kyokus []Kyoku `gorm:"many2many:artist_kyokus;" json:"kyokus"`
}
