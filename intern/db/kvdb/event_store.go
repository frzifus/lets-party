package kvdb

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"go.opentelemetry.io/otel/trace"

	"github.com/frzifus/lets-party/intern/model"
)

const bucketEvent = "event_store"

func NewEventStore(db *bolt.DB) (*EventStore, error) {
	const key = "event"
	return &EventStore{db: db, ekey: key}, db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketEvent))
		if err != nil {
			return err
		}
		j, err := json.Marshal(createDemoEvent())
		return bucket.Put([]byte(key), j)
	})
}

type EventStore struct {
	db   *bolt.DB
	ekey string
}

func (e *EventStore) GetEvent(ctx context.Context) (*model.Event, error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "GetEvent")
	defer span.End()

	span.AddEvent("View bucket")
	event := &model.Event{}
	return event, e.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketEvent))
		res := bucket.Get([]byte(e.ekey))
		if res == nil {
			err := errors.New("missing event")
			span.RecordError(err)
			return err
		}
		return json.Unmarshal(res, event)
	})
}

func (e *EventStore) UpdateEvent(ctx context.Context, in *model.Event) error {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "UpdateEvent")
	defer span.End()

	return e.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketEvent))
		event, err := json.Marshal(in)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(e.ekey), event)
	})
}

func createDemoEvent() *model.Event {
	return &model.Event{
		ID: uuid.MustParse("b0efa7fc-be99-4f5b-9fe8-1cd6cf6dd443"),
		Location: &model.Location{
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
