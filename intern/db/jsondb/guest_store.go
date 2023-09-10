package jsondb

import (
	"context"

	"github.com/google/uuid"

	"github.com/frzifus/lets-party/intern/model"
)

func NewGuestStore(filename string) *GuestStore {
	return &GuestStore{}
}

type GuestStore struct{}

func (GuestStore) CreateGuest(context.Context, *model.Guest) (uuid.UUID, error) {
	return uuid.New(), nil
}
func (GuestStore) UpdateGuest(context.Context, *model.Guest) error {
	return nil
}
func (GuestStore) ListGuests(context.Context) ([]*model.Guest, error) {
	return nil, nil
}
func (GuestStore) GetGuestByID(context.Context, uuid.UUID) (*model.Guest, error) {
	return &model.Guest{
		ID:              uuid.MustParse("39a502ac-ba10-430d-99ac-e0955eccb73b"),
		Firstname:       "Moritz",
		Lastname:        "Fleck",
		Child:           true,
		DietaryCategory: model.DietaryCatagoryOmnivore,
	}, nil
}
