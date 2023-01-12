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
	// load environmental variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// get database credentials from environmental variables
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	// init connection to database
	connection, err := sql.Open("mysql", dbUser+":"+dbPass+"@/"+dbName+"?parseTime=true")
	if err != nil {
		panic("could not connect to database")
	}
	DB = connection

	// init database tables
	DB.Exec(`
		CREATE TABLE IF NOT EXISTS User
		(
			Id       INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
			Username VARCHAR(60)                        NOT NULL,
			Email    VARCHAR(60)                        NOT NULL,
			Password VARCHAR(72)                        NOT NULL
		);
	`)

	DB.Exec(`
		ALTER TABLE User
    		ADD UNIQUE INDEX Email (Email);
	`)

	DB.Exec(`
		CREATE TABLE IF NOT EXISTS ActivityGroup
		(
			Id     INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
			Name   VARCHAR(25)                        NOT NULL,
			UserId INTEGER                            NOT NULL,
		
			FOREIGN KEY (UserId) REFERENCES User (Id)
				ON DELETE CASCADE
		);		
	`)

	DB.Exec(`
		CREATE TABLE IF NOT EXISTS Activity
		(
			Id       INTEGER     NOT NULL PRIMARY KEY AUTO_INCREMENT,
			Title    VARCHAR(60) NOT NULL,
			Body     TEXT        NOT NULL,
			ClosedOn DATETIME,
			OpenedOn DATETIME    NOT NULL,
			Due      DATETIME,
			UserId   INTEGER     NOT NULL,
			GroupId  INTEGER,
			
		    FULLTEXT KEY (Title, Body),
		    
			FOREIGN KEY (UserId) REFERENCES User (Id)
				ON DELETE CASCADE,
		
			FOREIGN KEY (GroupId) REFERENCES ActivityGroup (Id)
				ON DELETE CASCADE
		);
	`)

	DB.Exec(`
		ALTER TABLE Activity
    		ADD INDEX Opened (OpenedOn);
	`)
}
