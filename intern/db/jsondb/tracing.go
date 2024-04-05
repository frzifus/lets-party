// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package jsondb

import "go.opentelemetry.io/otel"

var tracer = otel.GetTracerProvider().Tracer("github.com/frzifus/lets-party/intern/db/jsondb")
