package consumer

import (
	"context"
	"fmt"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gofiber/fiber/v2/log"
	"github.com/muazwzxv/kafka-consumer-worker/internal/config"
	"github.com/muazwzxv/kafka-consumer-worker/internal/consumer/streamHandler"
	"github.com/muazwzxv/kafka-consumer-worker/internal/repository"
)

type Consumer struct {
	subscriber  *kafka.Subscriber
	handlers    map[string]streamHandler.MessageHandler
	config      *config.Config
	wg          sync.WaitGroup
	cancelFuncs []context.CancelFunc
}

// HandlerDependencies contains all dependencies needed by stream handlers
type HandlerDependencies struct {
	UserRepo repository.UserRepository
}

// Init creates a consumer with handlers built from config and dependencies
func Init(cfg *config.Config, deps *HandlerDependencies) (*Consumer, error) {
	handlers := make(map[string]streamHandler.MessageHandler)

	if cfg.Streams.UserLifecycle.Enable {
		handler := streamHandler.NewUserLifecycleHandler(
			deps.UserRepo,
			cfg.Streams.UserLifecycle.Topic,
		)
		handlers[cfg.Streams.UserLifecycle.Topic] = handler
		log.Infof("consumer: registered handler for topic: %s", cfg.Streams.UserLifecycle.Topic)
	}

	// if cfg.Streams.OrderEvents.Enable {
	//     handler := streamHandler.NewOrderEventsHandler(deps.OrderRepo, cfg.Streams.OrderEvents.Topic)
	//     handlers[cfg.Streams.OrderEvents.Topic] = handler
	//     log.Infof("consumer: registered handler for topic: %s", cfg.Streams.OrderEvents.Topic)
	// }

	return new(cfg, handlers)
}

func new(cfg *config.Config, handlers map[string]streamHandler.MessageHandler) (*Consumer, error) {
	subscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       []string{cfg.Kafka.Broker},
			Unmarshaler:   kafka.DefaultMarshaler{},
			ConsumerGroup: cfg.Kafka.ConsumerGroup,
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka subscriber: %w", err)
	}

	return &Consumer{
		subscriber:  subscriber,
		handlers:    handlers,
		config:      cfg,
		cancelFuncs: make([]context.CancelFunc, 0),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	if len(c.handlers) == 0 {
		log.Info("consumer: no handlers registered, skipping consumer start")
		return nil
	}

	log.Infof("consumer: starting with %d handlers", len(c.handlers))

	for topic, handler := range c.handlers {
		log.Infof("consumer: subscribing to topic: %s", topic)

		messages, err := c.subscriber.Subscribe(ctx, topic)
		if err != nil {
			return fmt.Errorf("subscribe to topic %s: %w", topic, err)
		}

		// Start a goroutine for each topic
		c.wg.Add(1)
		topicCtx, cancel := context.WithCancel(ctx)
		c.cancelFuncs = append(c.cancelFuncs, cancel)

		go c.processMessages(topicCtx, topic, messages, handler)
	}

	log.Info("consumer: started successfully")
	return nil
}

func (c *Consumer) processMessages(
	ctx context.Context,
	topic string,
	messages <-chan *message.Message,
	handler streamHandler.MessageHandler,
) {
	defer c.wg.Done()

	log.Infof("consumer: processing messages for topic: %s", topic)

	for {
		select {
		case <-ctx.Done():
			log.Infof("consumer: stopping message processing for topic: %s", topic)
			return
		case msg, ok := <-messages:
			if !ok {
				log.Warnf("consumer: message channel closed for topic: %s", topic)
				return
			}

			if err := handler.Handle(ctx, msg); err != nil {
				log.Errorf("consumer: error processing message %s from topic %s: %v",
					msg.UUID, topic, err)
				// Drop message silently - no Nack
			}
		}
	}
}

func (c *Consumer) Shutdown(ctx context.Context) error {
	log.Info("consumer: shutting down...")

	// Cancel all topic processing goroutines
	for _, cancel := range c.cancelFuncs {
		cancel()
	}

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("consumer: all goroutines stopped")
	case <-ctx.Done():
		log.Warn("consumer: shutdown timeout exceeded")
	}

	if err := c.subscriber.Close(); err != nil {
		return fmt.Errorf("close subscriber: %w", err)
	}

	log.Info("consumer: shutdown complete")
	return nil
}
