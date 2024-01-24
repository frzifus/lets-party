package model

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID   `json:"id"`
	CreatedAt *time.Time  `json:"created_at"`
	UpdatedAt *time.Time  `json:"updated_at"`
	Location  *Location   `json:"location"`
	Date      time.Time   `json:"date"`
	Hotels    []*Location `json:"hotels,omitempty"`
	Airports  []*Location `json:"airports,omitempty"`
}

type Location struct {
	ID           uuid.UUID  `json:"id"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"omitempty,updated_at"`
	Name         string     `json:"name,omitempty"`
	URL          string     `json:"url,omitempty"`
	Country      string     `json:"country,omitempty"`
	City         string     `json:"city,omitempty"`
	ZipCode      string     `json:"zipcode,omitempty"`
	Street       string     `json:"street,omitempty"`
	StreetNumber string     `json:"street_number,omitempty"`
	Longitude    float64    `json:"longitude,omitempty"`
	Latitude     float64    `json:"latitude,omitempty"`
	Website      string     `json:"website,omitempty"`
}
