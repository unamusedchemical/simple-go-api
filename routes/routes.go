package routes

import (
	"awesomeProject/controllers"
	"github.com/gofiber/fiber/v2"
)

// init api routes
func Setup(app *fiber.App) {
	app.Post("/api/register", controllers.Register)
	app.Post("/api/login", controllers.Login)
	app.Get("/api/user", controllers.GetUser)
	app.Post("/api/user/update", controllers.UpdateUser)
	app.Get("/api/logout", controllers.Logout)
	app.Post("/api/user/delete", controllers.DeleteUser)
	app.Post("/api/activity/new", controllers.CreateActivity)
	app.Post("/api/activity/update", controllers.UpdateActivity)
	app.Post("/api/activity/:id/delete", controllers.DeleteActivity)
	app.Post("/api/activity/group", controllers.EditGroupActivity)
	app.Get("/api/activities", controllers.Activities)
	app.Post("/api/activity/:id/close", controllers.CloseActivity)
	app.Post("/api/group/new", controllers.CreateGroup)
	app.Post("/api/group/update", controllers.UpdateGroup)
	app.Post("/api/group/:id/delete", controllers.DeleteGroup)
	app.Get("/api/group/:id/get", controllers.GetGroup)
	app.Get("/api/groups", controllers.GetGroups)
}
