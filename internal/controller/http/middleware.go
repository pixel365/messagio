package http

import (
	"os"
	"strings"

	"github.com/brianvoe/sjwt"
	"github.com/gofiber/fiber/v2"
	"github.com/pixel365/messagio/internal/model"
)

func checkTokenMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		token := ctx.Get("Authorization", "")
		if !strings.HasPrefix(token, "Bearer ") {
			return ctx.SendStatus(fiber.StatusUnauthorized)
		}

		token = token[len("Bearer "):]
		secret := os.Getenv("TOKEN_SECRET")

		if isValid := sjwt.Verify(token, []byte(secret)); !isValid {
			return ctx.SendStatus(fiber.StatusUnauthorized)
		}

		claims, err := sjwt.Parse(token)
		if err != nil {
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		var cl model.Claims
		err = claims.ToStruct(&cl)
		if err != nil {
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		ctx.Locals(model.ClaimsLocal, &cl)

		return ctx.Next()
	}
}
