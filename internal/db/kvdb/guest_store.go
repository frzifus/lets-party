// Copyright (C) 2024 the quixsi maintainers
// See root-dir/LICENSE for more information

package kvdb

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"go.opentelemetry.io/otel/trace"

	"github.com/quixsi/core/internal/model"
)

const bucketGuest = "guest_store"

func NewGuestStore(db *bolt.DB) (*GuestStore, error) {
	return &GuestStore{db: db}, db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketGuest))
		return err
	})
}

type GuestStore struct {
	db *bolt.DB
}

func (g *GuestStore) CreateGuest(ctx context.Context, guest *model.Guest) (uuid.UUID, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "CreateGuest")
	defer span.End()

	if guest.ID == uuid.Nil {
		span.AddEvent("uuid is nil, generate a new a new id")
		guest.ID = uuid.New()
	}

	j, err := json.Marshal(guest)
	if err != nil {
		return uuid.Nil, err
	}

	span.AddEvent("Update bucket")
	return guest.ID, g.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketGuest)).Put(guest.ID[:], j)
	})
}

func (g *GuestStore) UpdateGuest(ctx context.Context, guest *model.Guest) error {
	var span trace.Span
	_, span = tracer.Start(ctx, "UpdateGuest")
	defer span.End()

	if guest.ID == uuid.Nil {
		err := errors.New("guest ID is required for updating")
		span.RecordError(err)
		return err
	}
	now := time.Now()
	guest.UpdatedAt = &now

	j, err := json.Marshal(guest)
	if err != nil {
		return err
	}

	span.AddEvent("Update bucket")
	return g.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketGuest))
		res := bucket.Get(guest.ID[:])
		if res == nil {
			return errors.New("guest not found")
		}
		return bucket.Put(guest.ID[:], j)
	})
}

func (g *GuestStore) DeleteGuest(ctx context.Context, guestID uuid.UUID) error {
	var span trace.Span
	_, span = tracer.Start(ctx, "DeleteGuest")
	defer span.End()

	if guestID == uuid.Nil {
		err := errors.New("guest ID is required for updating")
		span.RecordError(err)
		return err
	}
	span.AddEvent("Update bucket")
	return g.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketGuest))
		return bucket.Delete(guestID[:])
	})
}

func (g *GuestStore) ListGuests(ctx context.Context) ([]*model.Guest, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "ListGuests")
	defer span.End()

	span.AddEvent("View bucket")
	var guests []*model.Guest
	return guests, g.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketGuest))
		return bucket.ForEach(func(_, v []byte) error {
			guest := &model.Guest{}
			if err := json.Unmarshal(v, guest); err != nil {
				span.RecordError(err)
				return err
			}
			guests = append(guests, guest)
			return nil
		})
	})
}

func (g *GuestStore) GetGuestByID(ctx context.Context, guestID uuid.UUID) (*model.Guest, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "GetGuestByID")
	defer span.End()
	span.AddEvent("View bucket")
	guest := &model.Guest{}
	return guest, g.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketGuest))
		res := bucket.Get(guestID[:])
		if res == nil {
			err := errors.New("guest not found")
			span.RecordError(err)
			return err
		}
		return json.Unmarshal(res, &guest)
	})
}
