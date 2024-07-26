package http

import (
	"github.com/gofiber/fiber/v2"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pixel365/messagio/internal/model"
)

func makeMessage(ctx *fiber.Ctx, m *model.InncomingMessage) (*model.Message, error) {
	if !m.IsValid() {
		return nil, ctx.SendStatus(fiber.StatusBadRequest)
	}

	claims := ctx.Locals(model.ClaimsLocal).(*model.Claims)
	id, _ := gonanoid.New(nanoIdDefaultSize)
	message := &model.Message{
		InncomingMessage: model.InncomingMessage{
			To:   m.To,
			Text: m.Text,
		},
		NanoId: id,
		UserId: claims.UserId,
		Status: model.Pending,
	}

	return message, nil
}
