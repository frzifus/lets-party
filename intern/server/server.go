package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/frzifus/lets-party/intern/db"
	"github.com/frzifus/lets-party/intern/model"
	"github.com/frzifus/lets-party/intern/server/form"
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
	mux.Use(gin.Logger(), gin.Recovery(), otelgin.Middleware(s.serviceName))
	mux.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	mux.GET("/:uuid", form.NewProcessor(s.iStore, s.tStore, s.gStore).Render)

	// TODO
	mux.POST("/:uuid/submit", func(c *gin.Context) { c.String(http.StatusOK, "thanks!") })

	mux.PUT("/:uuid/guests", func(c *gin.Context) {
		if c.Request.Header.Get("Hx-Request") == "true" {
			uuid, err := s.gStore.CreateGuest(c, &model.Guest{})
			if err != nil {
				panic("Could not create guest")
			}
	
			form.NewProcessor(s.iStore, s.tStore, s.gStore).RenderGuestInputBlock(c, uuid)
			return
		}

		// TODO: create guest with data from body
		c.String(http.StatusOK, "did not create user")
	})

	mux.GET("/:uuid/guests", func(c *gin.Context) { c.String(http.StatusOK, "thanks!") })

	mux.ServeHTTP(w, r)
}
