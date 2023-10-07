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
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/frzifus/lets-party/intern/db/jsondb"
	"github.com/frzifus/lets-party/intern/model"
	"github.com/frzifus/lets-party/intern/server"
)

func main() {
	var (
		serviceName = flag.String("service-name", "party-invite", "otel service name")
		addr        = flag.String("addr", "0.0.0.0:8080", "default server address")
		otlpAddr    = flag.String("otlp-grpc", "", "default otlp/gRPC address, by default disabled. Example value: localhost:4317")
		logLevelArg = flag.String("log-level", "INFO", "log level")
	)
	flag.Parse()
	fmt.Println("logLevel", *logLevelArg)
	logLevel := new(slog.Level)
	if err := logLevel.UnmarshalText([]byte(*logLevelArg)); err != nil {
		panic(err)
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)

	logger.Info("start and listen", "address", addr)
	logger.Info("otlp/gRPC", "address", otlpAddr, "service", serviceName)

	if *otlpAddr != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		grpcOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock()}
		conn, err := grpc.DialContext(ctx, *otlpAddr, grpcOptions...)
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
	}

	guestsStore, _ := jsondb.NewGuestStore("testdata/guests.json")

	_, _ = guestsStore.CreateGuest(context.Background(), &model.Guest{
		ID:              uuid.MustParse("39a502ac-ba10-430d-99ac-e0955eccb73b"),
		Firstname:       "Moritz",
		Lastname:        "Fleck",
		Child:           true,
		DietaryCategory: model.DietaryCatagoryOmnivore,
	})

	guests, _ := guestsStore.ListGuests(context.Background())
	for i, g := range guests {
		logger.Debug("guests", "number", i, "firstname", g.Firstname, "lastname", g.Lastname)
	}

	translationStore, _ := jsondb.NewTranslationStore("testdata/translations.json")
	invitationStore, _ := jsondb.NewInvitationStore("testdata/invitations.json")
	invitations, _ := invitationStore.ListInvitations(context.Background())
	for i, invite := range invitations {
		logger.Debug("invitations", "number", i, "inviteID", invite.ID, "guestIDs", invite.GuestIDs)
	}

	logger.Info("stats", "invitations", len(invitations), "guests", len(guests))
	srv := &http.Server{
		Addr: *addr,
		Handler: server.NewServer(
			*serviceName,
			invitationStore,
			guestsStore,
			translationStore,
		),
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("error during listen and serve", "error", err)
		os.Exit(1)
	}
	logger.Info("shutdown")
}
