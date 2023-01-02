package models

import "time"

var DataTimeFormat = "YYYY-MM-DD hh:mm:ss"

type Activity struct {
	Id              uint       `json:"id" gorm:"primaryKey;not null;autoIncrement;type:Integer"`
	ActivityName    string     `json:"activity_name" gorm:"not null; type:VARCHAR(30);"`
	ActivityContent string     `json:"activity-content" gorm:"not null;type:TEXT"`
	ClosedOn        *time.Time `json:"closed_on" gorm:"type: DATETIME"`
	OpenedOn        time.Time  `json:"opened_on" json:"OpenedOn" gorm:"not null; type: DATETIME"`
	Due             *time.Time `json:"due" gorm:"type: DATETIME"`
	UserId          uint       `json:"-" json:"UserId" gorm:"not null;type:INTEGER"`
	User            User       `json:"-" gorm:"foreignKey:UserId;references:Id;constraint:OnDelete:CASCADE"`
	LabelId         *uint      `json:"-" gorm:"type:INTEGER"`
	Label           Label      `json:"-" gorm:"foreignKey:LabelId;references:Id;constraint:OnDelete:SET NULL"`
}
