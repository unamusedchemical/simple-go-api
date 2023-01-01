package database

import (
	"awesomeProject/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	connection, err := gorm.Open(mysql.Open("root:alekochoveko@/Todo"), &gorm.Config{})
	if err != nil {
		panic("could not connect to database")
	}

	DB = connection

	connection.AutoMigrate(&models.User{})
	connection.AutoMigrate(&models.Activity{})
	connection.AutoMigrate(&models.Label{})
}
