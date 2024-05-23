// Copyright (C) 2024 the quixsi maintainers
// See root-dir/LICENSE for more information

package jsondb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	"github.com/quixsi/core/internal/model"
)

func NewInvitationStore(filename string) (*InvitationStore, error) {
	store := &InvitationStore{
		invitations: make(map[uuid.UUID][]uuid.UUID),
		filename:    filename,
	}

	if err := store.loadFromFile(); err != nil {
		return nil, err
	}
	return store, nil
}

type InvitationStore struct {
	mu          sync.RWMutex
	invitations map[uuid.UUID][]uuid.UUID
	filename    string
}

func (i *InvitationStore) GetInvitationByID(ctx context.Context, inviteID uuid.UUID) (*model.Invitation, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "GetInvitationByID")
	defer span.End()

	span.AddEvent("RLock")
	i.mu.RLock()
	defer span.AddEvent("RUnlock")
	defer i.mu.RUnlock()

	guestIDs, ok := i.invitations[inviteID]
	if !ok {
		err := fmt.Errorf("could not find invite with id: %s", inviteID)
		span.RecordError(err)
		return nil, err
	}
	return &model.Invitation{
		ID:       inviteID,
		GuestIDs: guestIDs,
	}, nil
}

func (i *InvitationStore) CreateInvitation(ctx context.Context, guestIDs ...uuid.UUID) (*model.Invitation, error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "CreateInvitation")
	defer span.End()

	span.AddEvent("Lock")
	i.mu.Lock()
	defer span.AddEvent("Unlock")
	defer i.mu.Unlock()
	id := uuid.New()
	if _, ok := i.invitations[id]; ok {
		err := fmt.Errorf("cannot create invitation, uuid already exists")
		span.RecordError(err)
		return nil, err
	}
	i.invitations[id] = guestIDs
	if err := i.saveToFile(ctx); err != nil {
		return nil, err
	}
	return &model.Invitation{
		ID:       id,
		GuestIDs: guestIDs,
	}, nil
}

func (i *InvitationStore) UpdateInvitation(ctx context.Context, invite *model.Invitation) error {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "UpdateInvitation")
	defer span.End()

	span.AddEvent("Lock")
	i.mu.Lock()
	defer span.AddEvent("Unlock")
	defer i.mu.Unlock()

	if _, ok := i.invitations[invite.ID]; !ok {
		err := fmt.Errorf("could not find invite")
		span.RecordError(err)
		return err
	}
	i.invitations[invite.ID] = invite.GuestIDs
	if err := i.saveToFile(ctx); err != nil {
		return err
	}
	return nil
}

func (i *InvitationStore) ListInvitations(ctx context.Context) ([]*model.Invitation, error) {
	var span trace.Span
	_, span = tracer.Start(ctx, "ListInvitations")
	defer span.End()

	span.AddEvent("RLock")
	i.mu.RLock()
	defer span.AddEvent("RUnlock")
	defer i.mu.RUnlock()

	var res []*model.Invitation
	for inviteID, guestIDs := range i.invitations {
		res = append(res, &model.Invitation{
			ID:       inviteID,
			GuestIDs: guestIDs,
		})
	}
	return res, nil
}

// saveToFile saves the current invitation store to the JSON file.
func (i *InvitationStore) saveToFile(ctx context.Context) error {
	var span trace.Span
	_, span = tracer.Start(ctx, "SaveToFile")
	defer span.End()

	fileData, err := json.MarshalIndent(i.invitations, "", "  ")
	if err != nil {
		span.RecordError(err)
		return err
	}

	err = os.WriteFile(i.filename, fileData, 0644)
	if err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// loadFromFile loads invitation data from the JSON file into the store.
func (i *InvitationStore) loadFromFile() error {
	if _, err := os.Stat(i.filename); os.IsNotExist(err) {
		// File does not exist, no invitations to load
		return nil
	}

	fileData, err := os.ReadFile(i.filename)
	if err != nil {
		return err
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	return json.Unmarshal(fileData, &i.invitations)
}
