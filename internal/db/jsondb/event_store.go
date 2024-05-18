// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package jsondb

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	"github.com/quixsi/core/internal/model"
)

func NewEventStore(filename string) (*EventStore, error) {
	store := &EventStore{
		filename: filename,
		event:    createDemoEvent(),
	}
	if err := store.loadFromFile(); err != nil {
		return nil, err
	}
	return store, nil
}

type EventStore struct {
	mu sync.RWMutex

	filename string
	event    *model.Event
}

func (e *EventStore) GetEvent(ctx context.Context) (*model.Event, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "GetEvent")
	defer span.End()

	span.AddEvent("Lock")
	e.mu.Lock()
	defer span.AddEvent("RUlock")
	defer e.mu.Unlock()

	return e.event, nil
}

func (e *EventStore) UpdateEvent(ctx context.Context, _ *model.Event) error {
	var span trace.Span
	_, span = tracer.Start(ctx, "UpdateEvent")
	defer span.End()

	span.AddEvent("RLock")
	e.mu.RLock()
	defer span.AddEvent("RUnlock")
	defer e.mu.RUnlock()

	err := errors.New("not implemented")
	span.RecordError(err)
	return err
}

func (e *EventStore) loadFromFile() error {
	if _, err := os.Stat(e.filename); os.IsNotExist(err) {
		// File does not exist, no guests to load
		return nil
	}

	fileData, err := os.ReadFile(e.filename)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	return json.Unmarshal(fileData, &e.event)
}

func createDemoEvent() *model.Event {
	return &model.Event{
		Location: &model.Location{
			ID:           uuid.MustParse("851ec3b7-f4ce-4319-96f9-67cc755b06ec"),
			Name:         "Party location",
			ZipCode:      "1337",
			Street:       "Milky Way",
			StreetNumber: "42",
			City:         "Somewhere",
			Country:      "Germany",
			Longitude:    106.6333,
			Latitude:     10.8167,
		},
		Hotels: []*model.Location{
			{
				ID:           uuid.MustParse("4e657dd1-2f75-48c7-ac87-1d3da0cc9b93"),
				Name:         "Demo Hotel 1",
				ZipCode:      "1337",
				Street:       "Milky Way",
				StreetNumber: "42",
				City:         "Somewhere",
				Country:      "Germany",
				Longitude:    106.6333,
				Latitude:     10.8167,
			},
		},
		Airports: []*model.Location{
			{
				ID:           uuid.MustParse("4716775f-575d-4524-a0bb-20630cb017b4"),
				Name:         "Demo Airport 1",
				ZipCode:      "1337",
				Street:       "Milky Way",
				StreetNumber: "42",
				City:         "Somewhere",
				Country:      "Germany",
				Longitude:    106.6333,
				Latitude:     10.8167,
			},
		},
		Date: time.Date(2023, 12, 24, 0, 0, 0, 0, time.Local),
	}
}
