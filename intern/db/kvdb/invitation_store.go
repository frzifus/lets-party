// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package kvdb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"go.opentelemetry.io/otel/trace"

	"github.com/quixsi/core/intern/model"
)

const bucketInvitation = "invitation_store"

func NewInvitationStore(db *bolt.DB) (*InvitationStore, error) {
	return &InvitationStore{db: db}, db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketInvitation))
		return err
	})
}

type InvitationStore struct {
	db *bolt.DB
}

func (i *InvitationStore) GetInvitationByID(ctx context.Context, inviteID uuid.UUID) (*model.Invitation, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "GetInvitationByID")
	defer span.End()

	span.AddEvent("RLock")
	defer span.AddEvent("RUnlock")

	invite := &model.Invitation{}
	return invite, i.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketInvitation))
		res := bucket.Get(inviteID[:])
		return json.Unmarshal(res, invite)
	})
}

func (i *InvitationStore) CreateInvitation(ctx context.Context, guestIDs ...uuid.UUID) (*model.Invitation, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "CreateInvitation")
	defer span.End()

	span.AddEvent("Lock")
	defer span.AddEvent("Unlock")
	id := uuid.New()
	invite := &model.Invitation{
		ID:       id,
		GuestIDs: guestIDs,
	}
	return invite, i.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketInvitation))
		res := bucket.Get(id[:])
		if res != nil {
			err := fmt.Errorf("cannot create invitation, uuid already exists")
			span.RecordError(err)
		}
		j, err := json.Marshal(invite)
		if err != nil {
			return err
		}
		return bucket.Put(id[:], j)
	})
}

func (i *InvitationStore) UpdateInvitation(ctx context.Context, invite *model.Invitation) error {
	var span trace.Span
	_, span = tracer.Start(ctx, "UpdateInvitation")
	defer span.End()

	span.AddEvent("Lock")
	defer span.AddEvent("Unlock")

	return i.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketInvitation))
		res := bucket.Get(invite.ID[:])
		if res == nil {
			err := fmt.Errorf("could not find invite")
			span.RecordError(err)
			return err
		}
		j, err := json.Marshal(invite)
		if err != nil {
			return err
		}
		return bucket.Put(invite.ID[:], j)
	})
}

func (i *InvitationStore) ListInvitations(ctx context.Context) ([]*model.Invitation, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "ListInvitations")
	defer span.End()

	span.AddEvent("RLock")
	defer span.AddEvent("RUnlock")

	var invites []*model.Invitation
	return invites, i.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketInvitation))
		return bucket.ForEach(func(_, v []byte) error {
			invite := &model.Invitation{}
			if err := json.Unmarshal(v, invite); err != nil {
				return err
			}
			invites = append(invites, invite)
			return nil
		})
	})
}
