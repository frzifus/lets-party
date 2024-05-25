// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/quixsi/core/internal/portal"
	templates "github.com/quixsi/core/internal/portal/tmp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var (
		addr        = flag.String("addr", "0.0.0.0:8080", "default server address")
		otlpAddr    = flag.String("otlp-grpc", "", "default otlp/gRPC address, by default disabled. Example value: localhost:4317")
		logLevelArg = flag.String("log-level", "INFO", "log level")
	)
	var logLevel slog.Level
	err := logLevel.UnmarshalText([]byte(*logLevelArg))
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(jsonHandler)
	if err != nil {
		logger.Error("unable to parse log level", "level-input", *logLevelArg, "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)
	logger.Info("log level set to", "log level", *logLevelArg)
	logger.Info("start and listen", "address", *addr)

	setupOTLP(*otlpAddr, logger)

	//NOTE: for implementation of observability (later)

	templateHandler := templates.NewTemplateHandler()

	portal := portal.NewPortal(logger, *addr, *templateHandler)
	portal.ServeHTTP()
}

func setupOTLP(otlpAddr string, logger *slog.Logger) {
	if otlpAddr != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		grpcOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock()}
		conn, err := grpc.DialContext(ctx, otlpAddr, grpcOptions...)
		if err != nil {
			logger.Error("failed to create gRPC connection to collector", "error", err)
			os.Exit(1)
		}
		defer conn.Close()

		// Set up a trace exporter
		otelExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			logger.Error("failed to create trace exporter", "error", err)
			os.Exit(1)
		}
		tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(otelExporter))
		otel.SetTracerProvider(tp)
	}
}
