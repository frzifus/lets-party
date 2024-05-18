package portal

import (
	"log/slog"
	"net/http"
	"os"

	templates "github.com/quixsi/core/internal/portal/tmp"
	sloghttp "github.com/samber/slog-http"
)

type Portal struct {
	logger  *slog.Logger
	address string
	routes  map[string]http.Handler
	//NOTE: just temporary until templ implementation
	templates templates.TemplateHandler
	//database interface{}
}

func NewPortal(
	logger *slog.Logger,
	address string,
	templates templates.TemplateHandler,
	//database interface{},
) *Portal {
	return &Portal{
		logger:    logger,
		address:   address,
		templates: templates,
		//database: nil,
	}
}

func (p *Portal) ServeHTTP() {
	mux := http.NewServeMux()

	loggerMW := sloghttp.NewWithConfig(
		p.logger, sloghttp.Config{
			DefaultLevel:     slog.LevelInfo,
			ClientErrorLevel: slog.LevelWarn,
			ServerErrorLevel: slog.LevelError,
			WithUserAgent:    true,
		},
	)

	p.routes = p.addRoutes()
	registerRoutes(mux, p.routes)

	portal := &http.Server{
		Addr:    p.address,
		Handler: loggerMW(mux),
	}

	p.logger.Info("listening on", "address", p.address)
	if err := portal.ListenAndServe(); err != nil {
		p.logger.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}
