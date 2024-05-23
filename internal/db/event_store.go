// Copyright (C) 2024 the quixsi maintainers
// See root-dir/LICENSE for more information

package db

import (
	"context"

	"github.com/quixsi/core/internal/model"
)

type EventStore interface {
	GetEvent(context.Context) (*model.Event, error)
	UpdateEvent(context.Context, *model.Event) error
}
