package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var DB *sql.DB

func Connect() {
	err := godotenv.Load("/Users/alekogeorgiev/GolandProjects/todo-fiber/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	connection, err := sql.Open("mysql", dbUser+":"+dbPass+"@/"+dbName+"?parseTime=true")
	if err != nil {
		panic("could not connect to database")
	}
	DB = connection

}
