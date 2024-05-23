// Copyright (C) 2024 the quixsi maintainers
// See root-dir/LICENSE for more information

package db

import (
	"context"

	"github.com/google/uuid"

	"github.com/quixsi/core/internal/model"
)

type GuestStore interface {
	CreateGuest(context.Context, *model.Guest) (uuid.UUID, error)
	UpdateGuest(context.Context, *model.Guest) error
	DeleteGuest(context.Context, uuid.UUID) error
	ListGuests(context.Context) ([]*model.Guest, error)
	GetGuestByID(context.Context, uuid.UUID) (*model.Guest, error)
}
