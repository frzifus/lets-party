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
	var span trace.Span
	ctx, span = tracer.Start(ctx, "CreateGuest")
	defer span.End()

	span.AddEvent("Lock")
	g.mu.Lock()
	defer span.AddEvent("Unlock")
	defer g.mu.Unlock()

	if guest.ID == uuid.Nil {
		guest.ID = uuid.New()
	}

	span.AddEvent("check if guest exists")
	// Add the guest to the store
	if _, ok := g.guests[guest.ID]; ok {
		err := errors.New("guest already exists")
		span.RecordError(err)
		return uuid.Nil, err
	}
	span.AddEvent("create new guest")
	now := time.Now()
	guest.CreatedAt = &now
	g.guests[guest.ID] = guest

	span.AddEvent("save to file")
	// Save the updated store to the JSON file
	if err := g.saveToFile(ctx); err != nil {
		return uuid.Nil, err
	}

	return guest.ID, nil
}

// UpdateGuest updates an existing guest's information in the store and JSON file.
func (g *GuestStore) UpdateGuest(ctx context.Context, guest *model.Guest) error {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "UpdateGuest")
	defer span.End()

	if guest.ID == uuid.Nil {
		err := errors.New("guest ID is required for updating")
		span.RecordError(err)
		return err
	}

	span.AddEvent("Lock")
	g.mu.Lock()
	defer span.AddEvent("Unlock")
	defer g.mu.Unlock()

	// Check if the guest exists in the store
	if _, ok := g.guests[guest.ID]; !ok {
		err := errors.New("guest not found")
		span.RecordError(err)
		return err
	}

	now := time.Now()
	guest.UpdatedAt = &now
	// Update the guest in the store
	g.guests[guest.ID] = guest

	// Save the updated store to the JSON file
	if err := g.saveToFile(ctx); err != nil {
		return err
	}

	return nil
}

// ListGuests returns a list of all guests in the store.
func (g *GuestStore) ListGuests(ctx context.Context) ([]*model.Guest, error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "ListGuests")
	defer span.End()

	span.AddEvent("Lock")
	g.mu.RLock()
	defer span.AddEvent("Unlock")
	defer g.mu.RUnlock()

	guestList := make([]*model.Guest, 0, len(g.guests))
	for _, guest := range g.guests {
		guestList = append(guestList, guest)
	}

	return guestList, nil
}

// GetGuestByID retrieves a guest by ID from the store.
func (g *GuestStore) GetGuestByID(ctx context.Context, id uuid.UUID) (*model.Guest, error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "GetGuestByID")
	defer span.End()

	span.AddEvent("RLock")
	g.mu.RLock()
	defer span.AddEvent("RUnlock")
	defer g.mu.RUnlock()

	guest, ok := g.guests[id]
	if !ok {
		err := errors.New("guest not found")
		span.RecordError(err)
		return nil, err
	}

	return guest, nil
}

// saveToFile saves the current guest store to the JSON file.
func (g *GuestStore) saveToFile(ctx context.Context) error {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "SaveToFile")
	defer span.End()

	fileData, err := json.MarshalIndent(g.guests, "", "  ")
	if err != nil {
		span.RecordError(err)
		return err
	}

	err = os.WriteFile(g.filename, fileData, 0644)
	if err != nil  {
		span.RecordError(err)
		return err
	}
	return nil
}

// loadFromFile loads guest data from the JSON file into the store.
func (g *GuestStore) loadFromFile() error {
	if _, err := os.Stat(g.filename); os.IsNotExist(err) {
		// File does not exist, no guests to load
		return nil
	}

	fileData, err := os.ReadFile(g.filename)
	if err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	return json.Unmarshal(fileData, &g.guests)
}

// DeleteGuest deletes an existing guest in the store and JSON file.
func (g *GuestStore) DeleteGuest(ctx context.Context, guestID uuid.UUID) error {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "DeleteGuest")
	defer span.End()

	if guestID == uuid.Nil {
		return errors.New("guest ID is required for updating")
	}

	span.AddEvent("Lock")
	g.mu.Lock()
	defer span.AddEvent("RUnlock")
	defer g.mu.Unlock()

	// Check if the guest exists in the store
	if _, ok := g.guests[guestID]; !ok {
		err := errors.New("guest not found")
		span.RecordError(err)
		return err
	}

	// Delete the guest from the store
	delete(g.guests, guestID)

	// Save the updated store to the JSON file
	if err := g.saveToFile(ctx); err != nil {
		return err
	}

	return nil
}
