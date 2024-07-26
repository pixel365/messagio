package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func NewRouter(app *fiber.App) {
	app.Get("/docs/*", swagger.HandlerDefault)
	app.Post("/auth", auth)

	api := app.Group("/api")
	api.Use(checkTokenMiddleware())
	api.Post("/single", single)
	api.Post("/batch", batch)
	api.Get("/details/:nanoid", details)
}
