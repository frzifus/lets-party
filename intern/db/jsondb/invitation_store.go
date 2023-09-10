package jsondb

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/frzifus/lets-party/intern/model"
)

func NewInvitationStore(filename string) *InvitationStore {
	id := "86761cef-5fd8-4849-b340-7c5e8ec0fed6"
	fmt.Println("[Debug] dummy invite ID:", id)
	return &InvitationStore{
		id: id,
	}
}

type InvitationStore struct {
	id string
}

func (i *InvitationStore) GetInvitationByID(context.Context, uuid.UUID) (*model.Invitation, error) {
	return &model.Invitation{
		ID:       uuid.MustParse(i.id),
		GuestIDs: []uuid.UUID{uuid.MustParse("39a502ac-ba10-430d-99ac-e0955eccb73b")},
	}, nil
}
