// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package main

import (
	"context"
	"log/slog"
	"os"

	bolt "go.etcd.io/bbolt"

	"github.com/quixsi/core/internal/db"
	"github.com/quixsi/core/internal/db/jsondb"
	"github.com/quixsi/core/internal/db/kvdb"
)

func main() {
	var (
	// inputType  = flag.String("input-type", "json", "")
	// inputPath  = flag.String("input-path", "json", "")
	// outputType = flag.String("output-type", "kvdb", "")
	// outputPath = flag.String("output-path", "json", "")
	)

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	logger := slog.New(jsonHandler)

	jdb := newJsonDB(logger, "../../testdata")
	kdb := newKVDB(logger, "output.db")
	logger.Info("start converting")
	into(kdb, jdb)
	logger.Info("finished converting")
}

type database interface {
	db.EventStore
	db.GuestStore
	db.InvitationStore
	db.TranslationStore
	Close() error
}

type dbWrapper struct {
	db.EventStore
	db.GuestStore
	db.InvitationStore
	db.TranslationStore

	closeFN func() error
}

func (d *dbWrapper) Close() error {
	return d.closeFN()
}

func into(dst, src database) {
	defer src.Close()
	defer dst.Close()
	ctx := context.Background()

	guests, err := src.ListGuests(ctx)
	if err != nil {
		panic(err)
	}
	for _, g := range guests {
		if _, err := dst.CreateGuest(ctx, g); err != nil {
			panic(err)
		}
	}
	invites, err := src.ListInvitations(ctx)
	if err != nil {
		panic(err)
	}
	for _, inv := range invites {
		if _, err := dst.CreateInvitation(ctx, inv.GuestIDs...); err != nil {
			panic(err)
		}
	}
	event, err := src.GetEvent(ctx)
	if err != nil {
		panic(err)
	}
	if err := dst.UpdateEvent(ctx, event); err != nil {
		panic(err)
	}
	list, err := src.ListLanguages(ctx)
	if err != nil {
		panic(err)
	}
	for _, key := range list {
		t, err := src.ByLanguage(ctx, key)
		if err != nil {
			panic(err)
		}
		if err := dst.CreateLanguage(ctx, key, t); err != nil {
			panic(err)
		}
	}
}

func newKVDB(logger *slog.Logger, path string) database {
	bdb, err := bolt.Open(path, 0600, nil)
	if err != nil {
		logger.Error("could not initialize guest store", "error", err)
		os.Exit(1)
	}

	guestsStore, err := kvdb.NewGuestStore(bdb)
	if err != nil {
		logger.Error("could not initialize guest bucket", "error", err)
		os.Exit(1)
	}

	invitationStore, err := kvdb.NewInvitationStore(bdb)
	if err != nil {
		logger.Error("could not initialize guest bucket", "error", err)
		os.Exit(1)
	}

	eventStore, err := kvdb.NewEventStore(bdb)
	if err != nil {
		logger.Error("could not initialize event bucket", "error", err)
		os.Exit(1)
	}

	translationStore, err := kvdb.NewTranslationStore(bdb)
	if err != nil {
		logger.Error("initialize translation bucket", "error", err)
	}

	return &dbWrapper{
		GuestStore:       guestsStore,
		TranslationStore: translationStore,
		InvitationStore:  invitationStore,
		EventStore:       eventStore,
		closeFN:          bdb.Close,
	}
}

func newJsonDB(logger *slog.Logger, path string) database {
	logger.Info("jsondb storage folder", "path", path)
	guestsStore, err := jsondb.NewGuestStore(path + "/guests.json")
	if err != nil {
		logger.Error("could not initialize guest store", "error", path)
		os.Exit(1)
	}
	translationStore, err := jsondb.NewTranslationStore(path + "/translations.json")
	if err != nil {
		logger.Error("could not initialize translation store", "error", path)
		os.Exit(1)
	}
	invitationStore, err := jsondb.NewInvitationStore(path + "/invitations.json")
	if err != nil {
		logger.Error("could not initialize invitation store", "error", path)
		os.Exit(1)
	}
	eventStore, err := jsondb.NewEventStore(path + "/event.json")
	if err != nil {
		logger.Error("could not initialize event store", "error", path)
		os.Exit(1)
	}
	return &dbWrapper{
		GuestStore:       guestsStore,
		TranslationStore: translationStore,
		InvitationStore:  invitationStore,
		EventStore:       eventStore,
		closeFN:          func() error { return nil },
	}
}
