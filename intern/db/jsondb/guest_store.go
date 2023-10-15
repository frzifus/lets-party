package jsondb

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/frzifus/lets-party/intern/model"
)

// GuestStore is an implementation of the GuestStore interface
// that stores guest data in a JSON file.
type GuestStore struct {
	filename string
	mu       sync.RWMutex
	guests   map[uuid.UUID]*model.Guest
}

// NewGuestStoreFile creates a new GuestStore instance.
func NewGuestStore(filename string) (*GuestStore, error) {
	store := &GuestStore{
		filename: filename,
		guests:   make(map[uuid.UUID]*model.Guest),
	}

	if err := store.loadFromFile(); err != nil {
		return nil, err
	}
	return store, nil
}

// CreateGuest adds a new guest to the store and stores it in the JSON file.
func (g *GuestStore) CreateGuest(ctx context.Context, guest *model.Guest) (uuid.UUID, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if guest.ID == uuid.Nil {
		guest.ID = uuid.New()
	}

	// Add the guest to the store
	if _, ok := g.guests[guest.ID]; ok {
		return uuid.Nil, errors.New("guest already exists")
	}
	now := time.Now()
	guest.CreatedAt = &now
	g.guests[guest.ID] = guest

	// Save the updated store to the JSON file
	if err := g.saveToFile(); err != nil {
		return uuid.Nil, err
	}

	return guest.ID, nil
}

// UpdateGuest updates an existing guest's information in the store and JSON file.
func (g *GuestStore) UpdateGuest(ctx context.Context, guest *model.Guest) error {
	if guest.ID == uuid.Nil {
		return errors.New("guest ID is required for updating")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Check if the guest exists in the store
	if _, ok := g.guests[guest.ID]; !ok {
		return errors.New("guest not found")
	}

	now := time.Now()
	guest.UpdatedAt = &now
	// Update the guest in the store
	g.guests[guest.ID] = guest

	// Save the updated store to the JSON file
	if err := g.saveToFile(); err != nil {
		return err
	}

	return nil
}

// ListGuests returns a list of all guests in the store.
func (g *GuestStore) ListGuests(ctx context.Context) ([]*model.Guest, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	guestList := make([]*model.Guest, 0, len(g.guests))
	for _, guest := range g.guests {
		guestList = append(guestList, guest)
	}

	return guestList, nil
}

// GetGuestByID retrieves a guest by ID from the store.
func (g *GuestStore) GetGuestByID(ctx context.Context, id uuid.UUID) (*model.Guest, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	guest, ok := g.guests[id]
	if !ok {
		return nil, errors.New("guest not found")
	}

	return guest, nil
}

// saveToFile saves the current guest store to the JSON file.
func (g *GuestStore) saveToFile() error {
	fileData, err := json.MarshalIndent(g.guests, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(g.filename, fileData, 0644)
}

// loadFromFile loads guest data from the JSON file into the store.
func (g *GuestStore) loadFromFile() error {
	if _, err := os.Stat(g.filename); os.IsNotExist(err) {
		// File does not exist, no guests to load
		return nil
	}

	fileData, err := ioutil.ReadFile(g.filename)
	if err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	return json.Unmarshal(fileData, &g.guests)
}

// DeleteGuest deletes an existing guest in the store and JSON file.
func (g *GuestStore) DeleteGuest(ctx context.Context, guestID uuid.UUID) error {
	if guestID == uuid.Nil {
		return errors.New("guest ID is required for updating")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Check if the guest exists in the store
	if _, ok := g.guests[guestID]; !ok {
		return errors.New("guest not found")
	}

	// Delete the guest from the store
	delete(g.guests, guestID)

	// Save the updated store to the JSON file
	if err := g.saveToFile(); err != nil {
		return err
	}

	return nil
}