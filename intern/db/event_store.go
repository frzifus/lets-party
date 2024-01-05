package db

import (
	"context"

	"github.com/frzifus/lets-party/intern/model"
)

type EventStore interface {
	GetEvent(context.Context) (*model.Event, error)
	UpdateEvent(context.Context, *model.Event) error
}
