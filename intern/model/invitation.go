package model

import "github.com/google/uuid"

type Invitation struct {
	ID       uuid.UUID
	GuestIDs []uuid.UUID
}
