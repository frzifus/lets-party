// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package model

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	*Location
	Date     time.Time   `json:"date" form:"date"`
	Hotels   []*Location `json:"hotels,omitempty" form:"hotels"`
	Airports []*Location `json:"airports,omitempty" form:"airports"`
}

type Location struct {
	ID           uuid.UUID  `json:"id" form:"-"`
	CreatedAt    *time.Time `json:"created_at" form:"-"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty" form:"-"`
	Name         string     `json:"name,omitempty" form:"name"`
	URL          string     `json:"url,omitempty" form:"url"`
	Country      string     `json:"country,omitempty" form:"country"`
	City         string     `json:"city,omitempty" form:"city"`
	ZipCode      string     `json:"zipcode,omitempty" form:"zipcode"`
	Street       string     `json:"street,omitempty" form:"street"`
	StreetNumber string     `json:"street_number,omitempty" form:"street_number"`
	Longitude    float64    `json:"longitude,omitempty" form:"longitude"`
	Latitude     float64    `json:"latitude,omitempty" form:"latitude"`
	Website      string     `json:"website,omitempty" form:"website"`
}
