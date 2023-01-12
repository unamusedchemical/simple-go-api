package main

import (
	"awesomeProject/database"
	"awesomeProject/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// init connection with database
	database.Connect()

	// when the main function ends - the connection closes
	defer database.DB.Close()

	// init web app
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))
	routes.Setup(app)

	// listen on localhost:8000
	app.Listen(":8000")
}
