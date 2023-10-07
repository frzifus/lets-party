package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/frzifus/lets-party/intern/db"
	"github.com/frzifus/lets-party/intern/server/templates"
)

func NewServer(
	serviceName string,
	iStore db.InvitationStore,
	gStore db.GuestStore,
	tStore db.TranslationStore,
) *Server {
	return &Server{
		serviceName: serviceName,
		iStore:      iStore,
		gStore:      gStore,
		tStore:      tStore,
	}
}

type Server struct {
	serviceName string
	iStore      db.InvitationStore
	gStore      db.GuestStore
	tStore      db.TranslationStore
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := gin.New()
	mux.Use(gin.Logger(), gin.Recovery(), inviteExists(s.iStore), otelgin.Middleware(s.serviceName))
	mux.NoRoute(notFound)

	guestHandler := templates.NewGuestHandler(s.iStore, s.tStore, s.gStore)
	mux.GET("/:uuid", guestHandler.RenderForm)
	mux.PUT("/:uuid/guests", guestHandler.Create)
	mux.DELETE("/:uuid/guests", guestHandler.Create)
	mux.POST("/:uuid/submit", guestHandler.Submit)
	// mux.PATCH("/:uuid/guests", guestHandler.Update)

	mux.GET("/:uuid/guests", func(c *gin.Context) { c.String(http.StatusOK, "thanks!") })

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
