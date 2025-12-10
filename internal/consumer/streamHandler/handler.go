package streamHandler

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type MessageHandler interface {
	Handle(ctx context.Context, msg *message.Message) error
	TopicName() string
}
