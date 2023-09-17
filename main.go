package main

import (
	"context"
	"flag"
	"fmt"

	"net/http"

	"github.com/google/uuid"

	"github.com/frzifus/lets-party/intern/db/jsondb"
	"github.com/frzifus/lets-party/intern/model"
	"github.com/frzifus/lets-party/intern/server"
)

func main() {
	var (
		serviceName = *flag.String("service-name", "party-invite", "otel service name")
		addr        = *flag.String("addr", "0.0.0.0:8080", "default server address")
	)

	guestsStore, _ := jsondb.NewGuestStore("testdata/guests.json")

	_, _ = guestsStore.CreateGuest(context.Background(), &model.Guest{
		ID:              uuid.MustParse("39a502ac-ba10-430d-99ac-e0955eccb73b"),
		Firstname:       "Moritz",
		Lastname:        "Fleck",
		Child:           true,
		DietaryCategory: model.DietaryCatagoryOmnivore,
	})

	guests, _ := guestsStore.ListGuests(context.Background())
	for _, g := range guests {
		fmt.Printf("Firstname %s, Lastname %s \n", g.Firstname, g.Lastname)
	}

	translationStore, _ := jsondb.NewTranslationStore("testdata/translations.json")
	invitationStore, _ := jsondb.NewInvitationStore("testdata/invitations.json")
	invitations, _ := invitationStore.ListInvitations(context.Background())
	for i, invite := range invitations {
		fmt.Printf("invitation(%d): %s with guests: %s", i, invite.ID, invite.GuestIDs)
	}

	srv := &http.Server{
		Addr: addr,
		Handler: server.NewServer(
			serviceName,
			invitationStore,
			guestsStore,
			translationStore,
		),
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
	fmt.Println("shutdown")
}
