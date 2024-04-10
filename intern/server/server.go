// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package server

import (
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"

	"github.com/frzifus/lets-party/intern/db"
	"github.com/frzifus/lets-party/intern/server/templates"
)

//go:embed all:static
var staticFS embed.FS

func NewServer(
	serviceName string,
	staticDir string,
	deadline time.Time,
	iStore db.InvitationStore,
	gStore db.GuestStore,
	tStore db.TranslationStore,
	eStore db.EventStore,
) *Server {
	return &Server{
		logger:      slog.Default().WithGroup("http"),
		serviceName: serviceName,
		staticDir:   staticDir,
		deadline:    deadline,
		iStore:      iStore,
		gStore:      gStore,
		tStore:      tStore,
		eStore:      eStore,
	}
}

type Server struct {
	serviceName string
	staticDir   string
	deadline    time.Time
	logger      *slog.Logger
	iStore      db.InvitationStore
	gStore      db.GuestStore
	tStore      db.TranslationStore
	eStore      db.EventStore
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := gin.New()
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	middlewares := []gin.HandlerFunc{
		sloggin.NewWithConfig(s.logger,
			sloggin.Config{
				DefaultLevel:     slog.LevelInfo,
				ClientErrorLevel: slog.LevelWarn,
				ServerErrorLevel: slog.LevelError,
			},
		),
		gin.Recovery(), otelgin.Middleware(s.serviceName), slogAddTraceAttributes,
	}

	username := "admin"
	if v, ok := os.LookupEnv("PARTY_ADMIN"); ok {
		username = v
	}

	password := "admin"
	if v, ok := os.LookupEnv("PARTY_PASSWORD"); ok {
		password = v
	}

	adminArea := mux.Group("/admin")
	adminArea.Use(append(middlewares, gin.BasicAuth(gin.Accounts{
		username: password,
	}))...)

	var staticDir fs.FS
	var err error
	switch {
	case s.staticDir != "":
		staticDir = os.DirFS(s.staticDir)
	default:
		staticDir, err = fs.Sub(staticFS, "static")
		if err != nil {
			panic(err)
		}
	}

	mux.StaticFS("/static", http.FS(fs.FS(staticDir)))

	if time.Now().After(s.deadline) {
		mux.Use(append(middlewares, readOnly(s.logger))...)
	}

	mux.Use(append(middlewares, inviteExists(s.iStore))...)
	guestHandler := templates.NewGuestHandler(s.iStore, s.tStore, s.gStore, s.eStore)
	mux.GET("/:uuid", guestHandler.RenderForm)
	mux.PUT("/:uuid/guests", guestHandler.Create)
	mux.DELETE("/:uuid/guests/:guestid", guestHandler.Delete)
	mux.POST("/:uuid/submit", guestHandler.Submit)

	adminArea.GET("/", guestHandler.RenderAdminOverview)
	adminArea.POST("/invitation", guestHandler.CreateInvitation)

	adminArea.POST("/event", guestHandler.UpdateEvent)
	adminArea.POST("/event/airports", guestHandler.CreateAirport)
	adminArea.DELETE("/event/airports/:uuid", guestHandler.DeleteAirport)
	adminArea.POST("/event/hotels", guestHandler.CreateHotel)
	adminArea.DELETE("/event/hotels/:uuid", guestHandler.DeleteHotel)

	translations := templates.NewTranslationHandler(s.tStore)
	adminArea.POST("/translations", translations.UpdateLanguage)

	mux.NoRoute(notFound)

	mux.ServeHTTP(w, r)
}

func inviteExists(db db.InvitationStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("uuid"))
		if err != nil {
			notFound(c)
			return
		}
		if _, err := db.GetInvitationByID(c.Request.Context(), id); err != nil {
			notFound(c)
			return
		}
		c.Next()
	}
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
}

func slogAddTraceAttributes(c *gin.Context) {
	sloggin.AddCustomAttributes(c,
		slog.String("trace-id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()),
	)
	sloggin.AddCustomAttributes(c,
		slog.String("span-id", trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()),
	)
	c.Next()
}

func readOnly(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			logger.ErrorContext(c.Request.Context(), "readOnly-mode", "error", errors.New("request method not allowed"))
			c.String(http.StatusMethodNotAllowed, "request method not allowed")
			c.Abort()
		}
		c.Next()
	}
}
