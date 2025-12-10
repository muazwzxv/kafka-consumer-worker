package streamHandler

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gofiber/fiber/v2/log"
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

	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Errorf("Failed to unmarshal message %s: %v", msg.UUID, err)
		return err
	}

	log.Infof("Message payload: %+v", payload)

	// TODO: Process with your repository
	// Example: Extract user data and call repository methods
	// if err := h.userRepo.Create(ctx, userData); err != nil {
	//     return err
	// }

	msg.Ack()
	return nil
}
