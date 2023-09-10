package db

import (
	"context"

	"github.com/google/uuid"

	"github.com/frzifus/lets-party/intern/model"
)

type GuestStore interface {
	CreateGuest(context.Context, *model.Guest) (uuid.UUID, error)
	UpdateGuest(context.Context, *model.Guest) error
	ListGuests(context.Context) ([]*model.Guest, error)
	GetGuestByID(context.Context, uuid.UUID) (*model.Guest, error)
}
