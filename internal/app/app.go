package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pixel365/messagio/internal/broker"
	"github.com/pixel365/messagio/internal/model"
	"github.com/pixel365/messagio/internal/postgres"
	"gorm.io/gorm"
)

var statuses = []model.Status{model.Completed, model.Error}

var application *Application

type Application struct {
	messages     chan *model.Message
	kafka        *broker.Broker
	Db           *gorm.DB
	isActive     bool
	MaxBatchSize int
}

func (a *Application) Cleanup() {
	a.isActive = false
	close(a.messages)
	a.kafka.Cleanup()

	if a.Db != nil {
		db, err := a.Db.DB()
		if err != nil {
			db.Close()
		}
	}
}

func (a *Application) NewMessage(message *model.Message) bool {
	if !message.IsValid() {
		return false
	}

	application.messages <- message

	return true
}

func Run(config *fiber.Config) *Application {
	if application == nil {
		application = &Application{
			isActive:     true,
			messages:     make(chan *model.Message),
			kafka:        broker.New(),
			Db:           postgres.New(),
			MaxBatchSize: 100,
		}

		application.Db.AutoMigrate(&model.Message{})

		go handleSignals()
		go handleKafkaEvents()

		for i := range application.kafka.TopicSegments {
			segment := i + 1
			go produceMessages(config, application.messages)

			for i := 0; i < application.kafka.TopicSegmentPartition; i++ {
				go consumeMessages(config, segment)
			}
		}
	}

	return application
}

func App() *Application {
	return application
}
