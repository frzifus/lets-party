// Copyright (C) 2024 the quixsi maintainers
// See root-dir/LICENSE for more information

package db

import (
	"context"

	"github.com/google/uuid"

	"github.com/quixsi/core/internal/model"
)

type InvitationStore interface {
	GetInvitationByID(context.Context, uuid.UUID) (*model.Invitation, error)
	UpdateInvitation(context.Context, *model.Invitation) error
	CreateInvitation(ctx context.Context, guestIDs ...uuid.UUID) (*model.Invitation, error)
	ListInvitations(ctx context.Context) ([]*model.Invitation, error)
}
