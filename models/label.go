package models

type Label struct {
	Id     uint   `json:"id" gorm:"primaryKey;not null; autoIncrement; type:INTEGER"`
	Name   string `json:"name" gorm:"type:VARCHAR(30); not null;class:FULLTEXT"`
	UserId uint   `json:"-" gorm:"type:INTEGER"`
	User   User   `json:"-" gorm:"foreignKey:UserId;references:Id;constraint:OnDelete:SET NULL"'`
}
