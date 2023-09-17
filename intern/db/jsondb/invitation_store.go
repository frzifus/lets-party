package jsondb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/google/uuid"

	"github.com/frzifus/lets-party/intern/model"
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

func (i *InvitationStore) GetInvitationByID(_ context.Context, inviteID uuid.UUID) (*model.Invitation, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	guestIDs, ok := i.invitations[inviteID]
	if !ok {
		return nil, fmt.Errorf("could not find invite with id: %s", inviteID)
	}
	return &model.Invitation{
		ID:       inviteID,
		GuestIDs: guestIDs,
	}, nil
}

func (i *InvitationStore) CreateInvitation(ctx context.Context, guestIDs ...uuid.UUID) (*model.Invitation, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	id := uuid.New()
	if _, ok := i.invitations[id]; ok {
		return nil, fmt.Errorf("cannot create invitation, uuid already exists")
	}
	i.invitations[id] = guestIDs
	if err := i.saveToFile(); err != nil {
		return nil, err
	}
	return &model.Invitation{
		ID:       id,
		GuestIDs: guestIDs,
	}, nil
}

func (i *InvitationStore) ListInvitations(ctx context.Context) ([]*model.Invitation, error) {
	i.mu.RLock()
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
func (i *InvitationStore) saveToFile() error {
	fileData, err := json.MarshalIndent(i.invitations, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(i.filename, fileData, 0644)
}

// loadFromFile loads invitation data from the JSON file into the store.
func (i *InvitationStore) loadFromFile() error {
	if _, err := os.Stat(i.filename); os.IsNotExist(err) {
		// File does not exist, no invitations to load
		return nil
	}

	fileData, err := ioutil.ReadFile(i.filename)
	if err != nil {
		return err
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	return json.Unmarshal(fileData, &i.invitations)
}
