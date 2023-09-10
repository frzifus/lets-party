package model

import (
	"time"

	"github.com/google/uuid"
)

type DietaryCategory int

const (
	DietaryCategoryUnknown DietaryCategory = iota
	DietaryCategoryVegan
	DietaryCategoryVegetarian
	DietaryCatagoryOmnivore
)

type InvitationStatus int

const (
	InvitationStatusUnknown InvitationStatus = iota
	InvitationStatusAccepted
	InvitationStatusRejected
	InvitationStatusNotAnswered
)

type Guest struct {
	ID               uuid.UUID
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
	Firstname        string
	Lastname         string
	Child            bool
	DietaryCategory  DietaryCategory
	InvitationStatus InvitationStatus
}
