package main

import (
	"flag"
	"fmt"

	"net/http"

	"github.com/frzifus/lets-party/intern/db/jsondb"
	"github.com/frzifus/lets-party/intern/server"
)

func main() {
	var (
		serviceName = *flag.String("service-name", "party-invite", "otel service name")
		addr        = *flag.String("addr", "0.0.0.0:8080", "default server address")
	)

	srv := &http.Server{
		Addr: addr,
		Handler: server.NewServer(
			serviceName,
			jsondb.NewInvitationStore("invites.json"),
			jsondb.NewGuestStore("guests.json"),
			jsondb.NewTranslationStore("translations.json"),
		),
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
	fmt.Println("shutdown")
}
