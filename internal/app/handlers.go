package app

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"

	k "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gofiber/fiber/v2"
	"github.com/pixel365/messagio/internal/model"
)

func handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)

	go func(app *Application, c chan os.Signal) {
		<-c
		app.Cleanup()
	}(App(), c)
}

func handleKafkaEvents() {
	for event := range application.kafka.DeliveryChannel {
		switch event.(type) {
		case *k.Message:
			slog.Info(event.String())
		case k.Error:
			slog.Error(event.String())
		}
	}
}

func consumeMessages(config *fiber.Config, segment int) {
	c := application.kafka.NewConsumer(segment)
	err := c.SubscribeTopics([]string{fmt.Sprintf("%s_%d", application.kafka.TopicPrefix, segment)}, nil)

	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	counter := 0
	for application.isActive {
		event := c.Poll(100)
		switch e := event.(type) {
		case *k.Message:
			var message model.Message
			err = config.JSONDecoder(e.Value, &message)
			if err == nil {
				if message.Status == model.Pending {
					status := statuses[rand.IntN(len(statuses))]
					application.Db.Model(&message).Where("nano_id = ?", message.NanoId).Update(
						"status", status,
					)
					counter += 1
				}
			}
			if counter%10 == 0 {
				go func(c *k.Consumer) {
					c.Commit()
				}(c)
			}
		case k.PartitionEOF:
			slog.Error(e.String())
		case k.Error:
			slog.Error(e.String())
		}
	}
}

func produceMessages(config *fiber.Config, c <-chan *model.Message) {
	for message := range c {
		result := application.Db.Create(message)
		if result.Error == nil {
			value, _ := config.JSONEncoder(message)

			// topic segmentation
			topic := fmt.Sprintf("%s_%d", application.kafka.TopicPrefix, 1)
			for i := range application.kafka.TopicSegments {
				segment := application.kafka.TopicSegments - i
				if int(message.ID)%segment == 0 {
					topic = fmt.Sprintf("%s_%d", application.kafka.TopicPrefix, segment)
					break
				}
			}

			// partition segmentation
			key := application.kafka.PartitionKeys[0]
			for i := range application.kafka.TopicSegmentPartition {
				segment := application.kafka.TopicSegmentPartition - i
				if int(message.ID)%segment == 0 {
					key = application.kafka.PartitionKeys[segment-1]
					break
				}
			}

			err := application.kafka.Producer.Produce(&k.Message{
				TopicPartition: k.TopicPartition{
					Topic:     &topic,
					Partition: k.PartitionAny,
				},
				Key:   []byte(key),
				Value: value,
			}, application.kafka.DeliveryChannel)

			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
