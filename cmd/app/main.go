package main

import (
	"log"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/pixel365/messagio/internal/app"
	_ "github.com/pixel365/messagio/internal/app/docs"
	"github.com/pixel365/messagio/internal/controller/http"
)

// @title Messagio API
// @version 1.0.1
// @description Test API service
// @termsOfService http://swagger.io/terms/
// @contact.name Ruslan Semagin
// @contact.email pixel365.sup@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host 62.109.20.78:8081
// @BasePath /
func main() {
	config := fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	}
	server := fiber.New(config)

	app.Run(&config)
	defer app.App().Cleanup()

	http.NewRouter(server)
	if err := server.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
