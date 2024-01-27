package kvdb

import "go.opentelemetry.io/otel"

var tracer = otel.GetTracerProvider().Tracer("github.com/frzifus/lets-party/intern/db/kvdb")
