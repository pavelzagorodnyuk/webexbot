package webexbot

import (
	"context"
)

type dialogProvider interface {
	provideFor(ctx context.Context, personId, roomId string) (Dialog, chan<- Event, context.CancelFunc)
}

type Dialog interface {
	context.Context

	PersonId() string
	RoomId() string
}
