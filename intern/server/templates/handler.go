package templates

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	"github.com/frzifus/lets-party/intern/db"
	"github.com/frzifus/lets-party/intern/model"
)

//go:embed *.html
var templates embed.FS

func NewGuestHandler(iStore db.InvitationStore, tStore db.TranslationStore, gStore db.GuestStore) *GuestHandler {
	coreTemplates := []string{"main.html", "footer.html"}
	invitationTemplates := []string{
		"invitation.header.html",
		"invitation.nav.html",
		"invitation.hero.html",
		"invitation.content.html",
		"greeting.html",
		"location.html",
		"date.html",
		"guest-form.html",
		"map.html",
	}
	languageTemplates := []string{"language.header.html", "language.content.html", "language-select.html"}

	return &GuestHandler{
		tmplForm: template.Must(template.ParseFS(templates, append(coreTemplates, invitationTemplates...)...)),
		tmplLang: template.Must(template.ParseFS(templates, append(coreTemplates, languageTemplates...)...)),
		iStore:   iStore,
		gStore:   gStore,
		tStore:   tStore,
		logger:   slog.Default().WithGroup("http"),
	}
}

type GuestHandler struct {
	tmplForm *template.Template
	tmplLang *template.Template
	iStore   db.InvitationStore
	gStore   db.GuestStore
	tStore   db.TranslationStore
	logger   *slog.Logger
}

func (p *GuestHandler) RenderAdminOverview(c *gin.Context) {
	c.String(http.StatusOK, "admin area!")
}

func (p *GuestHandler) RenderForm(c *gin.Context) {
	id := c.Param("uuid")
	uid, err := uuid.Parse(id)
	if err != nil {
		c.Error(err)
		return
	}

	invite, err := p.iStore.GetInvitationByID(c, uid)
	if err != nil {
		c.Error(err)
		return
	}

	lang := c.Query("lang")
	if lang == "" {
		langs, err := p.tStore.ListLanguages(c)
		languageOptions := make([]model.LanguageOption, len(langs))
		i := 0
		for _, lang := range langs {
			translation, err := p.tStore.ByLanguage(c, lang)
			if err != nil {
				panic(err)
			}
			languageOptions[i] = model.LanguageOption{
				Lang:       lang,
				FlagImgSrc: translation.FlagImgSrc,
			}
			i++
		}
		if err != nil {
			c.Error(err)
			return
		}
		if err := p.tmplLang.Execute(c.Writer, gin.H{"id": id, "languageOptions": languageOptions}); err != nil {
			c.Error(err)
		}
		return
	}

	translation, err := p.tStore.ByLanguage(c, lang)
	if err != nil {
		p.logger.ErrorContext(c.Request.Context(), "unknown target language", "error", err)
		c.String(http.StatusBadRequest, "unknown target language")
		return
	}

	var guests []*model.Guest
	for _, in := range invite.GuestIDs {
		g, err := p.gStore.GetGuestByID(c, in)
		if err != nil {
			c.Error(err)
			continue
		}
		guests = append(guests, g)
	}

	metadata := &model.Event{
		Location: &model.Location{
			ZipCode:      "1337",
			Street:       "Milky Way",
			StreetNumber: "42",
			City:         "Somewhere",
			Country:      "Germany",
			Longitude:    106.6333,
			Latitude:     10.8167,
		},
		Date: time.Date(2023, 12, 24, 0, 0, 0, 0, time.Local),
	}

	translation.Greeting, err = evalTemplate(translation.Greeting, guests)
	if err != nil {
		p.logger.ErrorContext(c.Request.Context(), "could not populate translation", "error", err)
		c.String(http.StatusInternalServerError, "could not render translation")
		return
	}

	p.tmplForm.Execute(c.Writer, gin.H{
		"id":          id,
		"metadata":    metadata,
		"translation": translation,
		"guests":      guests,
	})
}

func (p *GuestHandler) Submit(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		p.logger.ErrorContext(c.Request.Context(), "could not parse form", "error", err)
		c.String(http.StatusBadRequest, "could not parse form")
		return
	}

	for id, attrs := range p.parseGuestForm(c.Request.PostForm) {
		guestID, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		guest, err := p.gStore.GetGuestByID(c.Request.Context(), guestID)
		if err != nil {
			continue
		}
		guest.Parse(attrs)
		if err := p.gStore.UpdateGuest(c.Request.Context(), guest); err != nil {
			p.logger.ErrorContext(c.Request.Context(), "could update guest", "error", err)
		}
	}

	c.String(http.StatusOK, "thanks!")
}

// key: guestID
// val: from values
func (p *GuestHandler) parseGuestForm(raw url.Values) map[string]url.Values {
	input := make(map[string]url.Values)
	for k, v := range raw {
		got := strings.Split(k, ".")
		if len(got) != 2 {
			continue
		}
		if input[got[0]] == nil {
			input[got[0]] = make(url.Values)
		}
		input[got[0]][got[1]] = v
	}
	return input
}

func (p *GuestHandler) Create(c *gin.Context) {
	if c.Request.Header.Get("Hx-Request") == "true" {
		var span trace.Span
		ctx := c.Request.Context()
		ctx, span = tracer.Start(ctx, "GuestHandler.Create")
		defer span.End()

		inviteID, err := uuid.Parse(c.Param("uuid"))
		if err != nil {
			span.RecordError(err)
			p.logger.ErrorContext(ctx, "invalid inviteID", "error", err)
			c.String(http.StatusBadRequest, "invalid inviteID")
			return
		}

		invite, err := p.iStore.GetInvitationByID(ctx, inviteID)
		if err != nil {
			span.RecordError(err)
			p.logger.WarnContext(ctx, "invite not found", "error", err)
			c.String(http.StatusNotFound, "invite not found")
			return
		}

		gID, err := p.gStore.CreateGuest(ctx, &model.Guest{})
		if err != nil {
			span.RecordError(err)
			p.logger.ErrorContext(ctx, "could not create guest", "error", err)
			c.String(http.StatusBadRequest, "could not create guest")
			return
		}
		invite.GuestIDs = append(invite.GuestIDs, gID)
		if err := p.iStore.UpdateInvitation(ctx, invite); err != nil {
			span.RecordError(err)
			p.logger.WarnContext(ctx, "unable to update invite", "error", err)
			c.String(http.StatusInternalServerError, "unable to update invite")
			return
		}

		span.AddEvent("render guest input block")
		p.renderGuestInputBlock(ctx, c.Writer, c.DefaultQuery("lang", "en"), invite.ID, gID)
		return
	}

	// TODO: create guest with data from body
	c.String(http.StatusNotImplemented, "did not create user")
}

func (p *GuestHandler) Delete(c *gin.Context) {
	inviteID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		p.logger.ErrorContext(c.Request.Context(), "invalid invite ID", "error", err)
		c.String(http.StatusBadRequest, "invalid invite ID")
		return
	}
	guestID, err := uuid.Parse(c.Param("guestid"))
	if err != nil {
		p.logger.ErrorContext(c.Request.Context(), "invalid guest ID", "error", err)
		c.String(http.StatusBadRequest, "invalid guest ID")
		return
	}

	guest, err := p.gStore.GetGuestByID(c.Request.Context(), guestID)
	if err != nil {
		p.logger.ErrorContext(c.Request.Context(), "user not found", "error", err)
		c.String(http.StatusNotFound, "user not found")
		return
	}

	invite, err := p.iStore.GetInvitationByID(c.Request.Context(), inviteID)
	if err != nil {
		p.logger.ErrorContext(c.Request.Context(), "user does not belong to an invitation", "error", err)
		c.String(http.StatusNotFound, "user does not belong to an invitation")
		return
	}
	// TODO: tx
	invite.RemoveGuest(guest.ID)
	if err := p.iStore.UpdateInvitation(c.Request.Context(), invite); err != nil {
		p.logger.ErrorContext(c.Request.Context(), "unable to update invitation", "error", err)
		c.String(http.StatusInternalServerError, "unable to update invitation")
		return
	}

	if err := p.gStore.DeleteGuest(c.Request.Context(), guest.ID); err != nil {
		p.logger.ErrorContext(c.Request.Context(), "unable to delete guest", "error", err)
		c.String(http.StatusNotFound, "unable to delete guest")
		return
	}

	c.Status(http.StatusAccepted)
}

func (p *GuestHandler) Update(c *gin.Context) {
	c.String(http.StatusNotImplemented, "did not update user")
	// inviteID, err := uuid.Parse(c.Param("uuid"))
	// if err != nil {
	//	panic(err)
	// }
	// p.iStore.GetInvitationByID(c.Request.Context(), inviteID)
	// if err := p.gStore.UpdateGuest(c, &model.Guest{}); err != nil {
	//	c.String(http.StatusBadRequest, "could not update user")
	// }
	// c.String(http.StatusOK, "user update successful")
}

func (p *GuestHandler) renderGuestInputBlock(ctx context.Context, w gin.ResponseWriter, lang string, iID, gID uuid.UUID) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "GuestHandler.renderGuestInputBlock")
	defer span.End()

	translation, err := p.tStore.ByLanguage(ctx, lang)
	if err != nil {
		msg := "could not determine target language"
		span.AddEvent(msg)
		p.logger.WarnContext(ctx, msg, "error", err)
	}

	// TODO: Some options
	// - remove the wrapperTemplate, directly render guest-input and remove the define from guest-put
	// - make it possible to use the guest-input within the guest-form inside the guest-loop
	//	- this currently fails because without https://gohugo.io/functions/dict/ it seems it is not possible to pass both the $root data and the $guest data (".") to the template
	//	- missing $.translation data
	//	- https://stackoverflow.com/questions/18276173/calling-a-template-with-several-pipeline-parameters
	wrapperTemplate, _ := template.New("wrapper").Parse("{{ template \"GUEST_INPUT\" .}}")
	t, err := wrapperTemplate.ParseFS(templates, "guest-input.html")
	if err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "unable to parse guest input template", "error", err)
		return
	}

	err = t.Execute(w, gin.H{
		"invitationID": iID,
		"ID":           gID,
		"translation":  translation,
	})
	if err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "unable to render guest input template", "error", err)
		return
	}
}

func evalTemplate(msg string, data any) (string, error) {
	t, err := template.New("tmp").Parse(msg)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
