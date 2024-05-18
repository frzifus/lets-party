// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package model

import "github.com/google/uuid"

type Invitation struct {
	ID       uuid.UUID
	GuestIDs []uuid.UUID
}

func (i *Invitation) RemoveGuest(id uuid.UUID) {
	for idx, gid := range i.GuestIDs {
		if id == gid {
			i.GuestIDs = append(i.GuestIDs[:idx], i.GuestIDs[idx+1:]...)
			break
		}
	}
}
