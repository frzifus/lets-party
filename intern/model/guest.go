package model

import (
	"github.com/google/uuid"
)

type DietaryCategory int

const (
	DietaryCategoryUnknown DietaryCategory = iota
	DietaryCategoryVegan
	DietaryCategoryVegetarian
	DietaryCatagoryOmnivore
)

type Guest struct {
	ID              uuid.UUID
	Firstname       string
	Lastname        string
	Child           bool
	DietaryCategory DietaryCategory
}
