package logic

import (
	"context"

	"github.com/gofiber/fiber/v2/log"
	"github.com/muazwzxv/kafka-consumer-worker/internal/dto/stream"
	"github.com/muazwzxv/kafka-consumer-worker/internal/repository"
)

type UserCreatedLogic struct {
	UserRepo repository.UserRepository
}

func (l *UserCreatedLogic) ProcessUserPendingCreation(ctx context.Context, msg *stream.UserLifeCycleStream) error {
	user, err := l.UserRepo.GetByUUID(ctx, msg.UUID)
	if err != nil {
		return err
	}

	user.MarkAsActive()
	if updateErr := l.UserRepo.UpdateUser(ctx, user); updateErr != nil {
		return err
	}
	log.WithContext(ctx).Infof("successfully processed user pending creation event for uuid: %s", msg.UUID)

	return nil
}
