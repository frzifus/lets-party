package server

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

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
	iStore db.InvitationStore,
	gStore db.GuestStore,
	tStore db.TranslationStore,
	eStore db.EventStore,
) *Server {
	return &Server{
		logger:      slog.Default().WithGroup("http"),
		serviceName: serviceName,
		iStore:      iStore,
		gStore:      gStore,
		tStore:      tStore,
		eStore:      eStore,
	}
}

type Server struct {
	serviceName string
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

	adminArea := mux.Group("/admin")
	adminArea.Use(append(middlewares, gin.BasicAuth(gin.Accounts{
		"admin": "admin", // TODO: read from config, env variable...
	}))...)

	staticDir, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	mux.StaticFS("/static", http.FS(fs.FS(staticDir)))

	mux.Use(append(middlewares, inviteExists(s.iStore))...)
	guestHandler := templates.NewGuestHandler(s.iStore, s.tStore, s.gStore, s.eStore)
	mux.GET("/:uuid", guestHandler.RenderForm)
	mux.PUT("/:uuid/guests", guestHandler.Create)
	mux.DELETE("/:uuid/guests/:guestid", guestHandler.Delete)
	mux.POST("/:uuid/submit", guestHandler.Submit)

	adminArea.GET("/", guestHandler.RenderAdminOverview)
	adminArea.POST("/invitation", guestHandler.CreateInvitation)

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
