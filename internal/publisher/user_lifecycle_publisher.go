package publisher

import (
	"github.com/muazwzxv/kafka-consumer-worker/internal/config"
	"github.com/samber/do/v2"
)

var (
	userlifeCyclePublisherName = "user-lifecycle"
)

type UserLifecyclePublisher struct {
	Publisher
}

func NewUserLifecyclePublisher(i do.Injector) (*UserLifecyclePublisher, error) {
	cfg := do.MustInvoke[*config.Config](i)

	publisherConfig := cfg.Publishers.UserLifecycle
	if !publisherConfig.Enable {
		return &UserLifecyclePublisher{newNoopPublisher(publisherConfig.Topic)}, nil
	}

	pub, err := newPublisher(
		cfg.Kafka.Broker,
		cfg.Publishers.UserLifecycle.Topic,
		userlifeCyclePublisherName,
	)
	if err != nil {
		return nil, err
	}

	return &UserLifecyclePublisher{pub}, nil
}
