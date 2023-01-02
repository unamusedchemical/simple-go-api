package database

import (
	"awesomeProject/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	connection, err := gorm.Open(mysql.Open("root:alekochoveko@/todo-app?parseTime=true"), &gorm.Config{})
	if err != nil {
		panic("could not connect to database")
	}

	DB = connection

	connection.AutoMigrate(&models.User{})
	connection.AutoMigrate(&models.Activity{})
	connection.Exec("ALTER TABLE activities ADD FULLTEXT (activity_name)")
	connection.AutoMigrate(&models.Label{})
	connection.Exec("ALTER TABLE labels ADD FULLTEXT (name)")
}
