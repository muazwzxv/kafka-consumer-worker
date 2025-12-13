package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gofiber/fiber/v2/log"
)

type Publisher interface {
	Publish(ctx context.Context, payload interface{}) error
	Close() error
}

type publisher struct {
	kafkaPublisher *kafka.Publisher
	topic          string
	name           string
}

func newPublisher(broker, topic, name string) (Publisher, error) {
	kafkaPublisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   []string{broker},
			Marshaler: kafka.DefaultMarshaler{},
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return nil, err
	}

	log.Infof("%s publisher: initialized for topic: %s", name, topic)

	return &publisher{
		kafkaPublisher: kafkaPublisher,
		topic:          topic,
		name:           name,
	}, nil
}

func (p *publisher) Publish(ctx context.Context, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		log.WithContext(ctx).Errorw("failed to marshal payload",
			"publisher", p.name,
			"topic", p.topic,
			"error", err)
		return fmt.Errorf("marshal payload: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), data)

	log.WithContext(ctx).Infow("publishing message",
		"publisher", p.name,
		"topic", p.topic,
		"uuid", msg.UUID)

	if err := p.kafkaPublisher.Publish(p.topic, msg); err != nil {
		log.WithContext(ctx).Errorw("failed to publish message",
			"publisher", p.name,
			"topic", p.topic,
			"uuid", msg.UUID,
			"error", err)
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}

func (p *publisher) Close() error {
	log.Infof("%s publisher: closing", p.name)
	if err := p.kafkaPublisher.Close(); err != nil {
		log.Errorw("publisher close error",
			"publisher", p.name,
			"error", err)
		return err
	}

	log.Infof("%s publisher: closed", p.name)
	return nil
}

type noopPublisher struct {
	name string
}

func newNoopPublisher(name string) Publisher {
	log.Infof("%s publisher: disabled", name)
	return &noopPublisher{name: name}
}

func (n *noopPublisher) Publish(ctx context.Context, payload interface{}) error {
	return nil
}

func (n *noopPublisher) Close() error {
	return nil
}
