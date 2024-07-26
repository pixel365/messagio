package http

import (
	"math"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"github.com/brianvoe/sjwt"
	"github.com/gofiber/fiber/v2"
	"github.com/pixel365/messagio/internal/app"
	"github.com/pixel365/messagio/internal/model"
)

const (
	nanoIdDefaultSize = 21
	chunkSize         = 10
)

// @Summary      Authentication
// @Description  get access token
// @Accept       json
// @Produce      json
// @Param        creds body model.Creds true "Username (email) and password (min 6 characters)"
// @Success      200 {object} model.Token
// @Failure      400
// @Failure      422
// @Router       /auth [post]
func auth(ctx *fiber.Ctx) error {
	creds := new(model.Creds)
	if err := ctx.BodyParser(creds); err != nil {
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	if !creds.IsValid() {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	// find user in db
	userId := 1 // hardcoded fake user id just for example

	now := time.Now().UTC()
	expiresAt := now.Add(time.Hour * 24)

	claims := sjwt.New()
	claims.Set("salt", rand.IntN(1_000))
	claims.Set("username", creds.Username)
	claims.Set("user_id", userId)
	claims.Set("created_at", now.String())
	claims.SetExpiresAt(expiresAt)

	secretKey := []byte(os.Getenv("TOKEN_SECRET"))
	jwt := claims.Generate(secretKey)

	token := model.Token{
		AccessToken: jwt,
		TokenType:   "bearer",
		ExpiresAt:   expiresAt.String(),
	}

	return ctx.JSON(token)
}

// @Summary      Single message
// @Description  send single message
// @Accept       json
// @Produce      json
// @Param 		 Authorization header string true "Your access token" default(Bearer XXX)
// @Param        message body model.InncomingMessage true "Message object"
// @Success      200 {object} model.Message
// @Failure      400
// @Failure      401
// @Failure      422
// @Router       /api/single [post]
func single(ctx *fiber.Ctx) error {
	incomingMessage := new(model.InncomingMessage)
	if err := ctx.BodyParser(incomingMessage); err != nil {
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	message, err := makeMessage(ctx, incomingMessage)
	if err != nil {
		return err
	}

	if !app.App().NewMessage(message) {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	return ctx.JSON(&message)
}

// @Summary      Multiple messages
// @Description  send multiple messages
// @Accept       json
// @Produce      json
// @Param 		 Authorization header string true "Your access token" default(Bearer XXX)
// @Param        messages body []model.InncomingMessage true "Messages"
// @Success      200 {object} model.Message
// @Failure      400
// @Failure      401
// @Failure      422
// @Router       /api/batch [post]
func batch(ctx *fiber.Ctx) error {
	var messages []*model.InncomingMessage
	if err := ctx.BodyParser(&messages); err != nil {
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	total := len(messages)
	if total == 0 {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	batchResult := model.BatchResult{
		Encoder: ctx.App().Config().JSONEncoder,
	}

	if total > app.App().MaxBatchSize {
		batchResult.Ignored = total - app.App().MaxBatchSize
		total = app.App().MaxBatchSize
		messages = messages[:app.App().MaxBatchSize]
	}

	var wg sync.WaitGroup

	if total%10 != 0 {
		capacity := int(math.Ceil(float64(total)/10) * 10)
		for i := 0; i < (capacity - total); i++ {
			messages = append(messages, nil)
		}
	}

	chunks := math.Ceil(float64(total) / chunkSize)
	for i := range int(chunks) {
		j := i * chunkSize
		wg.Add(1)
		go func(wg *sync.WaitGroup, messages []*model.InncomingMessage, batchResult *model.BatchResult) {
			defer wg.Done()
			for _, message := range messages {
				if message != nil {
					if message.IsValid() {
						batchResult.ValidInc()
						m, err := makeMessage(ctx, message)
						if err == nil {
							if app.App().NewMessage(m) {
								batchResult.Result = append(batchResult.Result, *m)
							}
						}
					} else {
						batchResult.InvalidInc()
					}
				}
			}
		}(&wg, messages[j:j+chunkSize], &batchResult)
	}

	wg.Wait()

	return ctx.JSON(&batchResult)
}

// @Summary      Multiple messages
// @Description  send multiple messages
// @Accept       json
// @Produce      json
// @Param 		 Authorization header string true "Your access token" default(Bearer XXX)
// @Param        id path string true "Message ID"
// @Success      200 {object} model.Message
// @Failure      400
// @Failure      401
// @Failure      404
// @Failure      422
// @Router       /api/details/{id} [get]
func details(ctx *fiber.Ctx) error {
	id := ctx.Params("nanoid", "")
	if len(id) != nanoIdDefaultSize {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	claims := ctx.Locals(model.ClaimsLocal).(*model.Claims)

	var message model.Message
	app.App().Db.First(&message, "nano_id = ? AND user_id = ?", id, claims.UserId)

	if message.ID == 0 {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	return ctx.JSON(&message)
}
