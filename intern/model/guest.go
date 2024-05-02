// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

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

type GuestAgeCategory int

const (
	GuestAgeCategoryUnknown GuestAgeCategory = iota
	GuestAgeCategoryBaby
	GuestAgeCategoryTeenager
	GuestAgeCategoryAdult
)

type Guest struct {
	ID               uuid.UUID        `json:"id" form:"-"`
	Deleteable       bool             `json:"deleteable" form:"-"`
	CreatedAt        *time.Time       `json:"created_at" form:"-"`
	UpdatedAt        *time.Time       `json:"updated_at" form:"-"`
	Firstname        string           `json:"firstname" form:"firstname"`
	Lastname         string           `json:"lastname" form:"lastname"`
	AgeCategory      GuestAgeCategory `json:"age_category" form:"age_category"`
	DietaryCategory  DietaryCategory  `json:"dietary_category" form:"dietary_category"`
	InvitationStatus InvitationStatus `json:"invitation_status" form:"invitation_status"`
}
