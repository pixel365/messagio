package broker

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var instance *Broker

type Broker struct {
	DeliveryChannel       chan kafka.Event
	Producer              *kafka.Producer
	TopicPrefix           string
	Consumers             []*kafka.Consumer
	PartitionKeys         []string
	TopicSegments         int
	TopicSegmentPartition int
}

func (o *Broker) Cleanup() {
	close(o.DeliveryChannel)
	if o.Producer != nil {
		if !o.Producer.IsClosed() {
			o.Producer.Close()
		}
	}

	for _, c := range o.Consumers {
		if !c.IsClosed() {
			c.Close()
		}
	}
}

func (o *Broker) NewConsumer(segment int) *kafka.Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  os.Getenv("KAFKA_SERVER"),
		"group.id":           fmt.Sprintf("%s_%d", "consumers_group", segment),
		"auto.offset.reset":  kafka.OffsetBeginning,
		"enable.auto.commit": false,
	})
	if err != nil {
		log.Fatal(err)
	}

	o.Consumers = append(o.Consumers, c)

	return c
}

func New() *Broker {
	if instance == nil {
		topicSegments, err := strconv.Atoi(os.Getenv("TOPIC_SEGMENTS"))
		if err != nil || topicSegments < 1 {
			topicSegments = 3
		}

		topicSegmentPartition, err := strconv.Atoi(os.Getenv("TOPIC_SEGMENT_PARTITION"))
		if err != nil || topicSegmentPartition < 1 {
			topicSegmentPartition = 3
		}

		topicPrefix := os.Getenv("TOPIC_PREFIX")
		if topicPrefix == "" {
			topicPrefix = "new_messages"
		}

		instance = &Broker{
			DeliveryChannel:       make(chan kafka.Event),
			Producer:              newProducer(),
			TopicPrefix:           topicPrefix,
			TopicSegments:         topicSegments,
			TopicSegmentPartition: topicSegmentPartition,
		}

		for i := 0; i < instance.TopicSegmentPartition; i++ {
			instance.PartitionKeys = append(instance.PartitionKeys, fmt.Sprintf("%s_%d", "", i+1))
		}
	}

	return instance
}

func newProducer() *kafka.Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("KAFKA_SERVER"),
		"acks":              "all",
	})
	if err != nil {
		log.Fatal(err)
	}

	go func(p *kafka.Producer) {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Successfully produced record to topic %s partition [%d] @ offset %v\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			}
		}
	}(p)

	return p
}
