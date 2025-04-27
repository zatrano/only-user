package routes

import (
	handlers "zatrano/handlers/dashboard"
	"zatrano/middlewares"
	"zatrano/models"

	"github.com/gofiber/fiber/v2"
)

func registerDashboardRoutes(app *fiber.App) {
	dashboardGroup := app.Group("/dashboard")
	dashboardGroup.Use(
		middlewares.AuthMiddleware,
		middlewares.StatusMiddleware,
		middlewares.TypeMiddleware(models.System),
	)

	homeHandler := handlers.NewHomeHandler()
	dashboardGroup.Get("/home", homeHandler.HomePage)

	userHandler := handlers.NewUserHandler()
	dashboardGroup.Get("/users", userHandler.ListUsers)
	dashboardGroup.Get("/users/create", userHandler.ShowCreateUser)
	dashboardGroup.Post("/users/create", userHandler.CreateUser)
	dashboardGroup.Get("/users/update/:id", userHandler.ShowUpdateUser)
	dashboardGroup.Post("/users/update/:id", userHandler.UpdateUser)
	dashboardGroup.Post("/users/delete/:id", userHandler.DeleteUser)
	dashboardGroup.Delete("/users/delete/:id", userHandler.DeleteUser)
}
