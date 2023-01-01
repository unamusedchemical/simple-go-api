package routes

import (
	"awesomeProject/controllers"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Post("/api/search", controllers.Search)
	app.Post("/api/register", controllers.Register)
	app.Post("/api/login", controllers.Login)
	app.Get("/api/user", controllers.User)
	app.Post("/api/user/update", controllers.UpdateUser)
	app.Get("/api/logout", controllers.Logout)
	app.Post("/api/activity/new", controllers.CreateActivity)
	app.Post("/api/activity/:id/update", controllers.UpdateActivity)
	app.Post("/api/activity/:id/delete", controllers.DeleteActivity)
	app.Get("/api/activities", controllers.Activities)
	app.Post("/api/activity/:id/close", controllers.CloseActivity)
}
