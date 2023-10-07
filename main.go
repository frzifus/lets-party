package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/frzifus/lets-party/intern/db/jsondb"
	"github.com/frzifus/lets-party/intern/model"
	"github.com/frzifus/lets-party/intern/server"
)

func main() {

	var (
		serviceName = *flag.String("service-name", "party-invite", "otel service name")
		addr        = *flag.String("addr", "0.0.0.0:8080", "default server address")
		otlpAddr = *flag.String("otlp-grpc", "localhost:4317", "default otlp/gRPC address")
	)
	textHandler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(textHandler)

	logger.Info("start and listen", "address", addr)
	logger.Info("otlp/gRPC", "address", otlpAddr, "service", serviceName)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	grpcOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock()}
	conn, err := grpc.DialContext(ctx, otlpAddr, grpcOptions...)
	if err != nil {
		logger.Error("failed to create gRPC connection to collector", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Set up a trace exporter
	otelExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		logger.Error("failed to create trace exporter", err)
		os.Exit(1)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(otelExporter))

	otel.SetTracerProvider(tp)

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

	logger.Info("stats", "invitations", len(invitations), "guests", len(guests))
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
