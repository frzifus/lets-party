package db

import (
	"context"

	"github.com/google/uuid"

	"github.com/frzifus/lets-party/intern/model"
)

type InvitationStore interface {
	GetInvitationByID(context.Context, uuid.UUID) (*model.Invitation, error)
}
