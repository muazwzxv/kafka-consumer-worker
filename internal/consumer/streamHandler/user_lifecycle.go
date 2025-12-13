package streamHandler

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gofiber/fiber/v2/log"
	"github.com/muazwzxv/kafka-consumer-worker/internal/consumer/streamHandler/logic"
	"github.com/muazwzxv/kafka-consumer-worker/internal/dto/stream"
	"github.com/muazwzxv/kafka-consumer-worker/internal/entity"
	"github.com/muazwzxv/kafka-consumer-worker/internal/repository"
)

type UserLifecycleHandler struct {
	userRepo repository.UserRepository
	topic    string
}

func NewUserLifecycleHandler(userRepo repository.UserRepository, topic string) *UserLifecycleHandler {
	return &UserLifecycleHandler{
		userRepo: userRepo,
		topic:    topic,
	}
}

func (h *UserLifecycleHandler) TopicName() string {
	return h.topic
}

func (h *UserLifecycleHandler) Handle(ctx context.Context, msg *message.Message) error {
	log.Infof("Processing message from topic %s: %s", h.topic, msg.UUID)

	var userLifecycleStream *stream.UserLifeCycleStream
	if err := json.Unmarshal(msg.Payload, &userLifecycleStream); err != nil {
		log.Errorf("Failed to unmarshal message %s: %v", msg.UUID, err)
		return err
	}

	log.Infof("Message payload: %+v", userLifecycleStream)

	if userLifecycleStream.Status == entity.UserStatusPending.String() {
		l := logic.UserCreatedLogic{
			UserRepo: h.userRepo,
		}

		if err := l.ProcessUserPendingCreation(ctx, userLifecycleStream); err != nil {
			log.Errorf("Failed to process user pending activation uuid:%s: %v", msg.UUID, err)
		}
	}

	msg.Ack()
	return nil
}
