package model

import (
	"sync/atomic"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/golang-module/carbon/v2"
	"gorm.io/gorm"
)

type Status string

const (
	Pending   Status = "pending"
	Error     Status = "error"
	Completed Status = "completed"

	ClaimsLocal = "claims"
)

type Validatable interface {
	IsValid() bool
}

type Creds struct {
	Username string `json:"username" form:"username" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required"`
}

type Claims struct {
	UserName  string `json:"username"`
	CreatedAt string `json:"created_at"`
	Salt      int    `json:"salt"`
	UserId    int    `json:"user_id"`
	ExpiresAt int64  `json:"exp"`
}

type Token struct {
	AccessToken string `json:"access_token" validate:"required"`
	TokenType   string `json:"token_type" validate:"required"`
	ExpiresAt   string `json:"expires_at" validate:"required"`
}

type InncomingMessage struct {
	To   string `json:"to" validate:"required,email"`
	Text string `json:"text" validate:"required"`
}

type Message struct {
	gorm.Model       `json:"-"`
	InncomingMessage `gorm:"embedded"`
	NanoId           string `json:"id" gorm:"index;unique" validate:"required,len=21"`
	Status           Status `json:"status" gorm:"index" validate:"required"`
	UserId           int    `json:"-" validate:"required,min=1"`
}

type BatchResult struct {
	Encoder utils.JSONMarshal `json:"-"`
	Result  []Message
	Ignored int
	Valid   atomic.Uint32
	Invalid atomic.Uint32
}

func (b *BatchResult) ValidInc() {
	b.Valid.Add(1)
}

func (b *BatchResult) InvalidInc() {
	b.Invalid.Add(1)
}

func (b *BatchResult) MarshalJSON() ([]byte, error) {
	return b.Encoder(struct {
		Result  []Message `json:"result"`
		Ignored int       `json:"ignored"`
		Valid   uint32    `json:"valid"`
		Invalid uint32    `json:"invalid"`
	}{
		Result:  b.Result,
		Ignored: b.Ignored,
		Valid:   b.Valid.Load(),
		Invalid: b.Invalid.Load(),
	})
}

func (c *Creds) IsValid() bool {
	if len(c.Password) < 6 {
		return false
	}
	return isValid(c)
}

func (t *Token) IsValid() bool {
	if isValid(t) {
		now := carbon.Now(carbon.UTC)
		return carbon.Parse(t.ExpiresAt, carbon.UTC).Compare(">=", now)
	}
	return false
}

func (m *InncomingMessage) IsValid() bool {
	return isValid(m)
}

func (m *Message) IsValid() bool {
	switch m.Status {
	case Pending, Completed, Error:
		return isValid(m)
	default:
		return false
	}
}

func isValid(o interface{}) bool {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(o)
	return err == nil
}
