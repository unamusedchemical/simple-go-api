package models

type User struct {
	Id             uint    `json:"id" gorm:"primaryKey;not null;autoIncrement;type:INTEGER"`
	Username       string  `json:"username" gorm:"not null;type:VARCHAR(30)"`
	Email          string  `json:"email" gorm:"not null;type:VARCHAR(320);unique"`
	Password       []byte  `json:"-" gorm:"not null;type:VARCHAR(72)"`
	ProfilePicture *string `json:"-" gorm:"type:VARCHAR(256)"`
}
