package model

import "github.com/google/uuid"

type Invitation struct {
	ID       uuid.UUID
	GuestIDs []uuid.UUID
}

func (i *Invitation) RemoveGuest(id uuid.UUID) {
	for idx, id := range i.GuestIDs {
		if id == id {
			i.GuestIDs = append(i.GuestIDs[:idx], i.GuestIDs[idx+1:]...)
			break
		}
	}
}
