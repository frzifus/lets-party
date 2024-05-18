// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package db

import (
	"context"

	"github.com/quixsi/core/intern/model"
)

type EventStore interface {
	GetEvent(context.Context) (*model.Event, error)
	UpdateEvent(context.Context, *model.Event) error
}
