// Copyright (C) 2024 the quixsi maintainers
// See root-dir/LICENSE for more information

package templates

import "go.opentelemetry.io/otel"

var tracer = otel.GetTracerProvider().Tracer("github.com/quixsi/core/internal/server/templates")
