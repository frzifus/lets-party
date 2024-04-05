package kvdb

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/frzifus/lets-party/intern/model"
)

const bucketEvent = "event_store"

func NewEventStore(db *bolt.DB) (*EventStore, error) {
	const key = "event"

	logger := slog.Default().WithGroup("kvdb")
	return &EventStore{db: db, ekey: key}, db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketEvent))
		if err != nil {
			return err
		}
		res := bucket.Get([]byte(key))
		if err := json.Unmarshal(res, &model.Event{}); err != nil {
			logger.Warn("could not unmarshal event, create a new one", "error", err.Error())
			j, err := json.Marshal(createDemoEvent())
			if err != nil {
				panic(err)
			}
			return bucket.Put([]byte(key), j)
		}
		return nil
	})
}

type EventStore struct {
	db   *bolt.DB
	ekey string
}

func (e *EventStore) GetEvent(ctx context.Context) (*model.Event, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "GetEvent")
	defer span.End()

	span.AddEvent("View bucket")
	event := &model.Event{}
	return event, e.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketEvent))
		res := bucket.Get([]byte(e.ekey))
		if res == nil {
			err := errors.New("missing event")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		return json.Unmarshal(res, event)
	})
}

func (e *EventStore) UpdateEvent(ctx context.Context, in *model.Event) error {
	var span trace.Span
	_, span = tracer.Start(ctx, "UpdateEvent")
	defer span.End()

	return e.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketEvent))
		event, err := json.Marshal(in)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		return bucket.Put([]byte(e.ekey), event)
	})
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
		Date: time.Date(2023, 12, 24, 0, 0, 0, 0, time.Local),
	}
}
