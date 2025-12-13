package message

import "time"

type UserLifecycleMessage struct {
	ID          uint64
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
