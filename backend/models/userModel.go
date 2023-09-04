package models

import (
	"time"

	"gorm.io/gorm"
)

type GCProfile struct {
	GClassroomID string `json:"id"`
	EmailAddress string `json:"emailAddress"`
	Name         struct {
		FullName string `json:"fullName"`
	} `json:"name"`
	PhotoUrl    string `json:"photoUrl"`
	Permissions []struct {
		Permission string `json:"permission"`
	} `json:"permissions"`
}

type User struct {
	gorm.Model

	GCUID        string    `gorm:"column:gc_user_id;not null;uniqueIndex"` // Google Classroom user ID
	Username     string    `gorm:"column:username;not null"`
	Email        string    `gorm:"column:email;not null"`
	Token        string    `gorm:"column:token;not null"`
	TokenExpiry  time.Time `gorm:"column:token_expiry;not null"`
	RefreshToken string    `gorm:"column:refresh_token; not null"`
	PhotoUrl     string    `gorm:"column:photo_url;not null"`

	Courses []Course `gorm:"foreignKey:UserGCID;references:GCUID"`
}
