// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"net/http"
	"net/url"

	bolt "go.etcd.io/bbolt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/quixsi/core/internal/db"
	"github.com/quixsi/core/internal/db/kvdb"
	"github.com/quixsi/core/internal/server"
)

func main() {
	var (
		serviceName = flag.String("service-name", "party-invite", "otel service name")
		addr        = flag.String("addr", "0.0.0.0:8080", "default server address")
		dbStr       = flag.String("db", "kvdb://testdata/test.db", "database connection string")
		otlpAddr    = flag.String("otlp-grpc", "", "default otlp/gRPC address, by default disabled. Example value: localhost:4317")
		logLevelArg = flag.String("log-level", "INFO", "log level")
		staticDir   = flag.String("static-dir", "", "path to static directory")
		deadline    = flag.String("deadline", "", "deadline in format: 01 May 24 10:00 CET")
	)
	flag.Parse()
	fmt.Println("logLevel", *logLevelArg)
	var logLevel slog.Level
	err := logLevel.UnmarshalText([]byte(*logLevelArg))
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(jsonHandler)
	if err != nil {
		logger.Error("unable to parse log level", "level-input", *logLevelArg, "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)
	logger.Info("start and listen", "address", addr)
	logger.Info("otlp/gRPC", "address", otlpAddr, "service", serviceName)
	logger.Info("static-dir", "directory", staticDir)

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

	var dline time.Time
	if *deadline != "" {
		var err error
		dline, err = time.Parse(time.RFC822, *deadline)
		logger.Info("deadline set to", "date", *deadline)
		if err != nil {
			logger.Error("failed to parse deadline", err)
			os.Exit(1)
		}
	}

	var (
		guestsStore      db.GuestStore
		invitationStore  db.InvitationStore
		eventStore       db.EventStore
		translationStore db.TranslationStore
	)

	u, err := url.Parse(*dbStr)
	if err != nil {
		logger.Error("unable to parse db connection string", "error", err)
		os.Exit(1)
	}

	switch u.Scheme {
	case "kvdb":
		path := u.Host + u.Path
		db, err := bolt.Open(path, 0600, nil)
		if err != nil {
			logger.Error("could not initialize guest store", "error", err)
			os.Exit(1)
		}
		defer db.Close()

		guestsStore, err = kvdb.NewGuestStore(db)
		if err != nil {
			logger.Error("could not initialize guest bucket", "error", err)
			os.Exit(1)
		}

		invitationStore, err = kvdb.NewInvitationStore(db)
		if err != nil {
			logger.Error("could not initialize guest bucket", "error", err)
			os.Exit(1)
		}

		eventStore, err = kvdb.NewEventStore(db)
		if err != nil {
			logger.Error("could not initialize event bucket", "error", err)
			os.Exit(1)
		}

		translationStore, err = kvdb.NewTranslationStore(db)
		if err != nil {
			logger.Error("initialize translation bucket", "error", err)
		}
	default:
		logger.Error("Unknown storage backend", "type", u.Scheme)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr: *addr,
		Handler: server.NewServer(
			*serviceName,
			*staticDir,
			dline,
			invitationStore,
			guestsStore,
			translationStore,
			eventStore,
		),
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("error during listen and serve", "error", err)
		os.Exit(1)
	}
	logger.Info("shutdown")
}
