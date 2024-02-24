package templates

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	"github.com/frzifus/lets-party/intern/db"
	"github.com/frzifus/lets-party/intern/model"
	"github.com/frzifus/lets-party/intern/parser/form"
)

//go:embed *.html
var templates embed.FS

func NewGuestHandler(
	iStore db.InvitationStore,
	tStore db.TranslationStore,
	gStore db.GuestStore,
	eStore db.EventStore,
) *GuestHandler {
	coreTemplates := []string{"main.html", "footer.html", "main.style.html"}
	adminTemplates := []string{
		"admin.header.html",
		"admin.nav.html",
		"admin.content.html",
		"admin.event.html",
		"admin.event.location.html",
	}
	invitationTemplates := []string{
		"invitation.banner.html",
		"invitation.header.html",
		"invitation.nav.html",
		"invitation.hero.html",
		"invitation.content.html",
		"greeting.html",
		"location.html",
		"date.html",
		"guest-form.html",
		"map.html",
		"hotels.html",
		"airports.html",
	}
	languageTemplates := []string{"language.header.html", "language.content.html", "language-select.html"}

	return &GuestHandler{
		tmplAdmin: template.Must(template.ParseFS(templates, append(coreTemplates, adminTemplates...)...)),
		tmplForm:  template.Must(template.ParseFS(templates, append(coreTemplates, invitationTemplates...)...)),
		tmplLang:  template.Must(template.ParseFS(templates, append(coreTemplates, languageTemplates...)...)),
		iStore:    iStore,
		gStore:    gStore,
		tStore:    tStore,
		eStore:    eStore,
		logger:    slog.Default().WithGroup("http"),
	}
}

type GuestHandler struct {
	tmplAdmin *template.Template
	tmplForm  *template.Template
	tmplLang  *template.Template
	iStore    db.InvitationStore
	gStore    db.GuestStore
	tStore    db.TranslationStore
	eStore    db.EventStore
	logger    *slog.Logger
}

func (p *GuestHandler) RenderAdminOverview(c *gin.Context) {
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.RenderAdminOverview")
	defer span.End()

	metadata, err := p.eStore.GetEvent(ctx)
	if err != nil {
		p.logger.ErrorContext(ctx, "could not find event", "error", err)
		c.String(http.StatusInternalServerError, "could not find event")
		return
	}

	invs, err := p.iStore.ListInvitations(ctx)
	if err != nil {
		p.logger.ErrorContext(ctx, "could not list invitations", "error", err)
		c.String(http.StatusBadRequest, "could not list invitations")
		return
	}

	status := struct {
		Pending  int
		Accepted int
		Rejected int
	}{}

	table := make(map[uuid.UUID][]*model.Guest, len(invs))

	for _, inv := range invs {
		for _, gID := range inv.GuestIDs {
			fmt.Println("request guest:", gID.String())
			guest, err := p.gStore.GetGuestByID(ctx, gID)
			if err != nil {
				p.logger.WarnContext(
					ctx,
					"could not read guest", "error", err, "id", gID.String(),
				)
				continue
			}
			switch guest.InvitationStatus {
			case model.InvitationStatusAccepted:
				status.Accepted += 1
			case model.InvitationStatusRejected:
				status.Rejected += 1
			default:
				status.Pending += 1
			}
			table[inv.ID] = append(table[inv.ID], guest)
		}
	}

	p.tmplAdmin.Execute(c.Writer, gin.H{
		"metadata": metadata,
		"table":    table,
		"status":   status,
	})
}

func (p *GuestHandler) RenderForm(c *gin.Context) {
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.Submit")
	defer span.End()

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
		p.logger.ErrorContext(ctx, "unknown target language", "error", err)
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

	metadata, err := p.eStore.GetEvent(ctx)
	if err != nil {
		p.logger.ErrorContext(ctx, "could not find event", "error", err)
		c.String(http.StatusInternalServerError, "could not find event")
		return
	}

	guestsGreetList := make([]struct{ Firstname string }, len(guests))
	for index, guest := range guests {
		guestsGreetList[index].Firstname = guest.Firstname
		if index < len(guests)-2 {
			guestsGreetList[index].Firstname = fmt.Sprintf("%s,", guest.Firstname)
		} else if index < len(guests)-1 {
			guestsGreetList[index].Firstname = fmt.Sprintf("%s %s", guest.Firstname, translation.And)
		}
	}

	translation.Greeting, err = evalTemplate(translation.Greeting, guestsGreetList)
	if err != nil {
		p.logger.ErrorContext(ctx, "could not populate translation", "error", err)
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
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.Submit")
	defer span.End()

	if err := c.Request.ParseForm(); err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "could not parse form", "error", err)
		c.String(http.StatusBadRequest, "could not parse form")
		return
	}

	for id, attrs := range p.parseForm(c.Request.PostForm) {
		guestID, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		guest, err := p.gStore.GetGuestByID(ctx, guestID)
		if err != nil {
			continue
		}

		if err := form.Unmarshal(attrs, guest); err != nil {
			span.RecordError(err)
			p.logger.ErrorContext(ctx, "could not parse guest", "error", err)
			c.String(http.StatusBadRequest, "could not parse guest")
			return
		}

		if err := p.gStore.UpdateGuest(ctx, guest); err != nil {
			p.logger.ErrorContext(ctx, "could update guest", "error", err)
		}
	}

	lang := c.Query("lang")
	translation, err := p.tStore.ByLanguage(c, lang)
	if err != nil {
		p.logger.ErrorContext(ctx, "unknown target language", "error", err)
		c.String(http.StatusBadRequest, "unknown target language")
		return
	}

	wrapperTemplate, _ := template.New("wrapper").Parse("{{ template \"TOAST_SUCESS\" .}}")
	t, err := wrapperTemplate.ParseFS(templates, "toast.success.html")
	if err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "unable to parse toast.success template", "error", err)
		return
	}

	err = t.Execute(c.Writer, gin.H{
		"Title":   translation.Success.Title,
		"Message": translation.GuestForm.MessageSubmitSuccess,
	})
	if err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "unable to execute toast.success template", "error", err)
		return
	}
}

// key: guestID
// val: from values
func (p *GuestHandler) parseForm(raw url.Values) map[string]url.Values {
	input := make(map[string]url.Values)
	for k, v := range raw {
		got := strings.Split(k, ".")
		if len(got) < 2 {
			continue
		}
		if input[got[0]] == nil {
			input[got[0]] = make(url.Values)
		}
		input[got[0]][got[1]] = v
	}
	return input
}

func (p *GuestHandler) CreateInvitation(c *gin.Context) {
	if c.Request.Header.Get("Hx-Request") != "true" {
		c.String(http.StatusNotImplemented, "did not create invite")
	}
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.CreateInvitation")
	defer span.End()

	invs, err := p.iStore.ListInvitations(ctx)
	if err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "could not list invitations", "error", err)
		c.String(http.StatusInternalServerError, "could not list invitations")
		return
	}
	if len(invs) >= 250 { // HACK
		err := errors.New("maximum number of invitations exceeded")
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "can not add more invitations to this event", "error", err)
		c.String(http.StatusForbidden, "can not add more invitations to this event")
		return
	}

	// NOTE(workaround): create empty guest so that invite overview page can be rendered.
	gID, err := p.gStore.CreateGuest(ctx, &model.Guest{})
	if err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "could not create guest", "error", err)
		c.String(http.StatusBadRequest, "could not create guest")
		return
	}

	invite, err := p.iStore.CreateInvitation(ctx, gID)
	if err != nil {
		span.RecordError(err)
		p.logger.WarnContext(ctx, "could not create invite", "error", err)
		c.String(http.StatusNotFound, "could not create invite")
		return
	}

	wrapperTemplate, _ := template.New("wrapper").Parse("{{ template \"ADMIN_TABLE_INVITATION_ROW\" .}}")
	t, err := wrapperTemplate.ParseFS(templates, "admin.invitation-table-row.html")
	if err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "unable to parse invitation-table-row template", "error", err)
		return
	}

	err = t.Execute(c.Writer, gin.H{
		"inviteId": invite.ID.String(),
	})
	if err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "unable to execute invitation-table-row template", "error", err)
		return
	}
	// c.String(http.StatusCreated, invite.ID.String())
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

		if len(invite.GuestIDs) >= 10 { // HACK
			err := errors.New("maximum number of guests exceeded")
			span.RecordError(err)
			p.logger.ErrorContext(ctx, "can not add more guests to invite", "error", err)
			c.String(http.StatusForbidden, "can not add more guests to invite")
			return
		}

		gID, err := p.gStore.CreateGuest(ctx, &model.Guest{Deleteable: true})
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
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.Delete")
	defer span.End()

	inviteID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		p.logger.ErrorContext(ctx, "invalid invite ID", "error", err)
		c.String(http.StatusBadRequest, "invalid invite ID")
		return
	}
	guestID, err := uuid.Parse(c.Param("guestid"))
	if err != nil {
		p.logger.ErrorContext(ctx, "invalid guest ID", "error", err)
		c.String(http.StatusBadRequest, "invalid guest ID")
		return
	}

	guest, err := p.gStore.GetGuestByID(ctx, guestID)
	if err != nil {
		p.logger.ErrorContext(ctx, "user not found", "error", err)
		c.String(http.StatusNotFound, "user not found")
		return
	}

	if !guest.Deleteable {
		p.logger.ErrorContext(ctx, "user can not be deleted")
		c.String(http.StatusForbidden, "user can not be deleted")
		return
	}

	invite, err := p.iStore.GetInvitationByID(ctx, inviteID)
	if err != nil {
		p.logger.ErrorContext(ctx, "user does not belong to an invitation", "error", err)
		c.String(http.StatusNotFound, "user does not belong to an invitation")
		return
	}
	// TODO: tx
	invite.RemoveGuest(guest.ID)
	if err := p.iStore.UpdateInvitation(ctx, invite); err != nil {
		p.logger.ErrorContext(ctx, "unable to update invitation", "error", err)
		c.String(http.StatusInternalServerError, "unable to update invitation")
		return
	}

	if err := p.gStore.DeleteGuest(ctx, guest.ID); err != nil {
		p.logger.ErrorContext(ctx, "unable to delete guest", "error", err)
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

func (p *GuestHandler) CreateAirport(c *gin.Context) {
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.CreateAirport")
	defer span.End()
	e, err := p.eStore.GetEvent(ctx)
	if err != nil {
		span.RecordError(err)
		return
	}
	e.Airports = append(e.Airports, &model.Location{})
	if err := p.eStore.UpdateEvent(ctx, e); err != nil {
		span.RecordError(err)
	}
}

func (p *GuestHandler) DeleteAirport(c *gin.Context) {
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.DeleteAirport")
	defer span.End()

	airportID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		span.RecordError(err)
		return
	}
	e, err := p.eStore.GetEvent(ctx)
	if err != nil {
		span.RecordError(err)
		return
	}

	for i := 0; i < len(e.Airports); i++ {
		if e.Airports[i].ID == airportID {
			e.Airports = append(e.Airports[:i], e.Airports[i+1:]...)
			break
		}
	}

	if err := p.eStore.UpdateEvent(ctx, e); err != nil {
		span.RecordError(err)
	}
}

func (p *GuestHandler) CreateHotel(c *gin.Context) {
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.CreateHotel")
	defer span.End()
	e, err := p.eStore.GetEvent(ctx)
	if err != nil {
		span.RecordError(err)
		return
	}
	e.Hotels = append(e.Hotels, &model.Location{})
	if err := p.eStore.UpdateEvent(ctx, e); err != nil {
		span.RecordError(err)
	}
}

func (p *GuestHandler) DeleteHotel(c *gin.Context) {
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.DeleteHotel")
	defer span.End()

	hotelID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		span.RecordError(err)
		return
	}
	e, err := p.eStore.GetEvent(ctx)
	if err != nil {
		span.RecordError(err)
		return
	}

	for i := 0; i < len(e.Hotels); i++ {
		if e.Hotels[i].ID == hotelID {
			e.Hotels = append(e.Hotels[:i], e.Hotels[i+1:]...)
			break
		}
	}

	if err := p.eStore.UpdateEvent(ctx, e); err != nil {
		span.RecordError(err)
	}
}

func (p *GuestHandler) UpdateEvent(c *gin.Context) {
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.UpdateEvent")
	defer span.End()
	e, err := p.eStore.GetEvent(ctx)
	if err != nil {
		span.RecordError(err)
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "could not parse form", "error", err)
		c.String(http.StatusBadRequest, "could not parse form")
		return
	}

	var eventData url.Values
	raw := p.parseForm(c.Request.PostForm)
	for k, v := range raw {
		if k == e.ID.String() {
			eventData = v
			delete(raw, k)
			break
		}
	}

	{ // TODO: remove 2nd form unmarshal
		if err := form.Unmarshal(eventData, e); err != nil {
			span.RecordError(err)
			p.logger.ErrorContext(ctx, "could not parse event date", "error", err)
			c.String(http.StatusBadRequest, "could not parse event date")
			return
		}
		// HACK: form unmarshal does not support embedded structs
		if err := form.Unmarshal(eventData, e.Location); err != nil {
			span.RecordError(err)
			p.logger.ErrorContext(ctx, "could not parse event location", "error", err)
			c.String(http.StatusBadRequest, "could not parse event location")
			return
		}
	}

	for id, ldata := range raw {
		ldata["id"] = []string{id}
		l := model.Location{}
		if err := form.Unmarshal(ldata, &l); err != nil {
			span.RecordError(err)
			p.logger.ErrorContext(ctx, "could not parse other location", "error", err)
			continue
		}
		for i := 0; i < len(e.Airports); i++ {
			if l.ID == e.Airports[i].ID {
				e.Airports[i] = &l
			}
		}
		for i := 0; i < len(e.Hotels); i++ {
			if l.ID == e.Hotels[i].ID {
				e.Hotels[i] = &l
			}
		}
	}

	if err := p.eStore.UpdateEvent(ctx, e); err != nil {
		span.RecordError(err)
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
