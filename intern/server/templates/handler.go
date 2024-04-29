// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package templates

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	txttemplate "text/template"
	"time"
	_ "time/tzdata"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jeremywohl/flatten/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
		"admin.event.location.airport.html",
		"admin.event.location.hotel.html",
		"admin.translations.html",
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
		// NOTE: workaround to allow html formatting
		tmplForm: txttemplate.Must(txttemplate.ParseFS(templates, append(coreTemplates, invitationTemplates...)...)),
		tmplLang: template.Must(template.ParseFS(templates, append(coreTemplates, languageTemplates...)...)),
		iStore:   iStore,
		gStore:   gStore,
		tStore:   tStore,
		eStore:   eStore,
		logger:   slog.Default().WithGroup("http"),
	}
}

type GuestHandler struct {
	tmplAdmin *template.Template
	tmplForm  *txttemplate.Template
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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		p.logger.ErrorContext(ctx, "could not find event", "error", err)
		c.String(http.StatusInternalServerError, "could not find event")
		return
	}

	langs, err := p.tStore.ListLanguages(c)
	translations := make(map[string]map[string]string)

	for _, lang := range langs {
		// TODO:: handle errors
		translation, _ := p.tStore.ByLanguage(ctx, lang)
		out, _ := json.Marshal(translation)
		flattened, _ := flatten.FlattenString(string(out), "", flatten.DotStyle)
		result := make(map[string]string)
		_ = json.Unmarshal([]byte(flattened), &result)
		translations[lang] = result
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		_ = c.Error(err)
		return
	}

	invs, err := p.iStore.ListInvitations(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not list invitations")
		p.logger.ErrorContext(ctx, "could not list invitations", "error", err)
		c.String(http.StatusBadRequest, "could not list invitations")
		return
	}

	status := struct {
		Invitations struct {
			Total    int
			Pending  int
			Accepted int
			Rejected int
		}
		Diet struct {
			Unknown    int
			Vegetarian int
			Vegan      int
			Omnivore   int
		}
		AgeCategory struct {
			Unknown  int
			Baby     int
			Teenager int
			Adult    int
		}
	}{}

	table := make(map[uuid.UUID][]*model.Guest, len(invs))

	for _, inv := range invs {
		for _, gID := range inv.GuestIDs {
			guest, err := p.gStore.GetGuestByID(ctx, gID)
			if err != nil {
				p.logger.WarnContext(
					ctx,
					"could not read guest", "error", err, "id", gID.String(),
				)
				continue
			}
			status.Invitations.Total++
			switch guest.InvitationStatus {
			case model.InvitationStatusAccepted:
				status.Invitations.Accepted += 1
				switch guest.DietaryCategory {
				case model.DietaryCategoryUnknown:
					status.Diet.Unknown += 1
				case model.DietaryCategoryVegan:
					status.Diet.Vegan += 1
				case model.DietaryCategoryVegetarian:
					status.Diet.Vegetarian += 1
				case model.DietaryCatagoryOmnivore:
					status.Diet.Omnivore += 1
				}
				switch guest.AgeCategory {
				case model.GuestAgeCategoryUnknown:
					status.AgeCategory.Unknown += 1
				case model.GuestAgeCategoryBaby:
					status.AgeCategory.Baby += 1
				case model.GuestAgeCategoryTeenager:
					status.AgeCategory.Teenager += 1
				case model.GuestAgeCategoryAdult:
					status.AgeCategory.Adult += 1
				}
			case model.InvitationStatusRejected:
				status.Invitations.Rejected += 1
			default:
				status.Invitations.Pending += 1
			}
			table[inv.ID] = append(table[inv.ID], guest)
		}
	}

	if err := p.tmplAdmin.Execute(c.Writer, gin.H{
		"metadata":     metadata,
		"table":        table,
		"status":       status,
		"translations": translations,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not exec admin template")
	}
}

func (p *GuestHandler) RenderForm(c *gin.Context) {
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.Submit")
	defer span.End()

	id := c.Param("uuid")
	uid, err := uuid.Parse(id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		_ = c.Error(err)
		return
	}

	invite, err := p.iStore.GetInvitationByID(c, uid)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		_ = c.Error(err)
		return
	}

	lang := c.Query("lang")
	if lang == "" {
		langs, err := p.tStore.ListLanguages(c)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			_ = c.Error(err)
			return
		}
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
		if err := p.tmplLang.Execute(c.Writer, gin.H{"id": id, "languageOptions": languageOptions}); err != nil {
			_ = c.Error(err)
		}
		return
	}

	translation, err := p.tStore.ByLanguage(c, lang)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unknown target language")
		p.logger.ErrorContext(ctx, "unknown target language", "error", err)
		c.String(http.StatusBadRequest, "unknown target language")
		return
	}

	var guests []*model.Guest
	for _, in := range invite.GuestIDs {
		g, err := p.gStore.GetGuestByID(c, in)
		if err != nil {
			_ = c.Error(err)
			span.RecordError(err)
			span.SetStatus(codes.Error, "unknown guest")
			continue
		}
		guests = append(guests, g)
	}

	metadata, err := p.eStore.GetEvent(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not find event")
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not render greeting")
		p.logger.ErrorContext(ctx, "could not render greeting", "error", err)
		c.String(http.StatusInternalServerError, "could not render greeting")
		return
	}

	cetLocation, err := time.LoadLocation("CET")
	if err != nil {
		panic(err)
	}

	helper := map[string]any{
		"newline":      "<br />",
		"bolt":         "<b>",
		"boltend":      "</b>",
		"locationname": metadata.Name,
		"partytimeCET": metadata.Date.In(cetLocation).Format("15:04 MST"),
		"partydate":    metadata.Date.In(cetLocation).Format("2. January"),
		"isSingular":   len(guests) == 1,
	}

	translation.WelcomeMessage, err = evalTemplateUnsafe(translation.WelcomeMessage, helper)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not render welcome message")
		p.logger.ErrorContext(ctx, "could not render welcome message", "error", err)
		c.String(http.StatusInternalServerError, "could not render welcome message")
		return
	}

	if err := p.tmplForm.Execute(c.Writer, gin.H{
		"id":          id,
		"metadata":    metadata,
		"translation": translation,
		"guests":      guests,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could exec form template")
		p.logger.ErrorContext(ctx, "could exec form template", "error", err)
		c.String(http.StatusInternalServerError, "could exec form template")
		return
	}
}

func (p *GuestHandler) Submit(c *gin.Context) {
	var span trace.Span
	ctx := c.Request.Context()
	ctx, span = tracer.Start(ctx, "GuestHandler.Submit")
	defer span.End()

	if err := c.Request.ParseForm(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not parse form")
		p.logger.ErrorContext(ctx, "could not parse form", "error", err)
		c.String(http.StatusBadRequest, "could not parse form")
		return
	}

	for id, attrs := range p.parseForm(c.Request.PostForm) {
		guestID, err := uuid.Parse(id)
		if err != nil {
			span.AddEvent("invalid guest ID")
			continue
		}
		guest, err := p.gStore.GetGuestByID(ctx, guestID)
		if err != nil {
			span.AddEvent("could not load guest")
			continue
		}

		if err := form.Unmarshal(attrs, guest); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "could not unmarshal guest")
			p.logger.ErrorContext(ctx, "could not unmarshal guest", "error", err)
			c.String(http.StatusBadRequest, "could not unmarshal guest")
			return
		}

		if err := p.gStore.UpdateGuest(ctx, guest); err != nil {
			p.logger.ErrorContext(ctx, "could update guest", "error", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, "could update guest")

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
		span.SetStatus(codes.Error, "unable to parse toast.success template")
		p.logger.ErrorContext(ctx, "unable to parse toast.success template", "error", err)
		return
	}

	err = t.Execute(c.Writer, gin.H{
		"Title":   translation.Success.Title,
		"Message": translation.GuestForm.MessageSubmitSuccess,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to execute toast.success template")
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
		span.SetStatus(codes.Error, "could not list invitations")
		p.logger.ErrorContext(ctx, "could not list invitations", "error", err)
		c.String(http.StatusInternalServerError, "could not list invitations")
		return
	}
	if len(invs) >= 250 { // HACK
		err := errors.New("maximum number of invitations exceeded")
		span.RecordError(err)
		span.SetStatus(codes.Error, "can not add more invitations to this event")
		p.logger.ErrorContext(ctx, "can not add more invitations to this event", "error", err)
		c.String(http.StatusForbidden, "can not add more invitations to this event")
		return
	}

	// NOTE(workaround): create empty guest so that invite overview page can be rendered.
	gID, err := p.gStore.CreateGuest(ctx, &model.Guest{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not create guest")
		p.logger.ErrorContext(ctx, "could not create guest", "error", err)
		c.String(http.StatusBadRequest, "could not create guest")
		return
	}

	invite, err := p.iStore.CreateInvitation(ctx, gID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not create invite")
		p.logger.WarnContext(ctx, "could not create invite", "error", err)
		c.String(http.StatusNotFound, "could not create invite")
		return
	}

	wrapperTemplate, _ := template.New("wrapper").Parse("{{ template \"ADMIN_TABLE_INVITATION_ROW\" .}}")
	t, err := wrapperTemplate.ParseFS(templates, "admin.invitation-table-row.html")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to parse invitation-table-row template")
		p.logger.ErrorContext(ctx, "unable to parse invitation-table-row template", "error", err)
		return
	}

	err = t.Execute(c.Writer, gin.H{
		"inviteId": invite.ID.String(),
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to execute invitation-table-row template")
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
			span.SetStatus(codes.Error, "invalid inviteID")
			p.logger.ErrorContext(ctx, "invalid inviteID", "error", err)
			c.String(http.StatusBadRequest, "invalid inviteID")
			return
		}

		invite, err := p.iStore.GetInvitationByID(ctx, inviteID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invite not found")
			p.logger.WarnContext(ctx, "invite not found", "error", err)
			c.String(http.StatusNotFound, "invite not found")
			return
		}

		if len(invite.GuestIDs) >= 10 { // HACK
			err := errors.New("maximum number of guests exceeded")
			span.RecordError(err)
			span.SetStatus(codes.Error, "can not add more guests to invite")
			p.logger.ErrorContext(ctx, "can not add more guests to invite", "error", err)
			c.String(http.StatusForbidden, "can not add more guests to invite")
			return
		}

		gID, err := p.gStore.CreateGuest(ctx, &model.Guest{Deleteable: true})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "could not create guest")
			p.logger.ErrorContext(ctx, "could not create guest", "error", err)
			c.String(http.StatusBadRequest, "could not create guest")
			return
		}
		invite.GuestIDs = append(invite.GuestIDs, gID)
		if err := p.iStore.UpdateInvitation(ctx, invite); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "unable to update invite")
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid invite ID")
		p.logger.ErrorContext(ctx, "invalid invite ID", "error", err)
		c.String(http.StatusBadRequest, "invalid invite ID")
		return
	}
	guestID, err := uuid.Parse(c.Param("guestid"))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid invite ID")
		p.logger.ErrorContext(ctx, "invalid guest ID", "error", err)
		c.String(http.StatusBadRequest, "invalid guest ID")
		return
	}

	guest, err := p.gStore.GetGuestByID(ctx, guestID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user not found")
		p.logger.ErrorContext(ctx, "user not found", "error", err)
		c.String(http.StatusNotFound, "user not found")
		return
	}

	if !guest.Deleteable {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user can not be deleted")
		p.logger.ErrorContext(ctx, "user can not be deleted")
		c.String(http.StatusForbidden, "user can not be deleted")
		return
	}

	invite, err := p.iStore.GetInvitationByID(ctx, inviteID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user does not belong to an invitation")
		p.logger.ErrorContext(ctx, "user does not belong to an invitation", "error", err)
		c.String(http.StatusNotFound, "user does not belong to an invitation")
		return
	}
	// TODO: tx
	invite.RemoveGuest(guest.ID)
	if err := p.iStore.UpdateInvitation(ctx, invite); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to update invitation")
		p.logger.ErrorContext(ctx, "unable to update invitation", "error", err)
		c.String(http.StatusInternalServerError, "unable to update invitation")
		return
	}

	if err := p.gStore.DeleteGuest(ctx, guest.ID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to delete guest")
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
		span.SetStatus(codes.Error, "unable to parse guest input template")
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
		span.SetStatus(codes.Error, "unable to render guest input template")
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
		span.SetStatus(codes.Error, err.Error())
		return
	}
	newAirport := &model.Location{ID: uuid.New()}
	e.Airports = append(e.Airports, newAirport)
	if err := p.eStore.UpdateEvent(ctx, e); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	wrapperTemplate, _ := template.New("wrapper").Parse("{{ template \"ADMIN_EVENT_LOCATION_AIRPORT\" .airport}}")
	t, err := wrapperTemplate.ParseFS(templates, "admin.event.location.html", "admin.event.location.airport.html")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to parse invitation-table-row template")
		p.logger.ErrorContext(ctx, "unable to parse invitation-table-row template", "error", err)
		return
	}

	err = t.Execute(c.Writer, gin.H{
		"airport": newAirport,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to execute invitation-table-row template")
		p.logger.ErrorContext(ctx, "unable to execute invitation-table-row template", "error", err)
		return
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
		span.SetStatus(codes.Error, err.Error())
		return
	}
	e, err := p.eStore.GetEvent(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
		span.SetStatus(codes.Error, err.Error())
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
		span.SetStatus(codes.Error, err.Error())
		return
	}
	newHotel := &model.Location{ID: uuid.New()}
	e.Hotels = append(e.Hotels, newHotel)
	if err := p.eStore.UpdateEvent(ctx, e); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	wrapperTemplate, _ := template.New("wrapper").Parse("{{ template \"ADMIN_EVENT_LOCATION_HOTEL\" .hotel}}")
	t, err := wrapperTemplate.ParseFS(templates, "admin.event.location.html", "admin.event.location.hotel.html")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to parse invitation-table-row template")
		p.logger.ErrorContext(ctx, "unable to parse invitation-table-row template", "error", err)
		return
	}

	err = t.Execute(c.Writer, gin.H{
		"hotel": newHotel,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to execute invitation-table-row template")
		p.logger.ErrorContext(ctx, "unable to execute invitation-table-row template", "error", err)
		return
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
		span.SetStatus(codes.Error, err.Error())
		return
	}
	e, err := p.eStore.GetEvent(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
		span.SetStatus(codes.Error, err.Error())
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
		span.SetStatus(codes.Error, "could not get event")
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not parse form")
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

	const layout = "2006-01-02 15:04:05 -0700 MST"
	dateStr, ok := eventData["date"]
	if ok && len(dateStr) == 1 {
		ts, err := time.Parse(layout, dateStr[0])
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "could not event date timestamp")
			p.logger.ErrorContext(ctx, "could not event date timestamp", "error", err)
			c.String(http.StatusBadRequest, "could not event date timestamp")
			return
		}
		e.Date = ts
	}

	if err := form.Unmarshal(eventData, e); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not parse event")
		p.logger.ErrorContext(ctx, "could not parse event", "error", err)
		c.String(http.StatusBadRequest, "could not parse event")
		return
	}

	for id, ldata := range raw {
		ldata["id"] = []string{id}
		l := model.Location{}
		if err := form.Unmarshal(ldata, &l); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "could not parse other location")
			p.logger.ErrorContext(ctx, "could not parse other location", "error", err)
			continue
		}
		lID, err := uuid.Parse(id)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid uuid")
			p.logger.ErrorContext(ctx, "invalid uuid", "error", err)
			continue
		}
		for i := 0; i < len(e.Airports); i++ {
			if lID == e.Airports[i].ID {
				e.Airports[i] = &l
			}
		}
		for i := 0; i < len(e.Hotels); i++ {
			if lID == e.Hotels[i].ID {
				e.Hotels[i] = &l
			}
		}
	}

	if err := p.eStore.UpdateEvent(ctx, e); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "could not update event")
	}
}

type TranslationHandler struct {
	tStore db.TranslationStore
}

func NewTranslationHandler(tStore db.TranslationStore) *TranslationHandler {
	return &TranslationHandler{tStore: tStore}
}

func (t *TranslationHandler) UpdateLanguage(c *gin.Context) {
	ctx, span := tracer.Start(c, "TranslationHandler.UpdateLanguages")
	defer span.End()

	if err := c.Request.ParseForm(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	span.AddEvent("Read form entries", trace.WithAttributes(attribute.Int("count", len(c.Request.Form))))

	const valueSep = "::"
	// NOTE: In the following section, suffix numbers of transferred keys are
	// removed and sorted into a list.
	//
	// e.g.:
	// request.Form:
	// en.optionX => d
	// en.optionY.2 => c
	// en.optionY.0 => a
	// en.optionY.1 => b
	//
	// formValues:
	// en.optionX => d
	// en.option> => [a,b,c]
	formValues := url.Values{}
	for key, value := range c.Request.Form {
		kk := strings.Split(key, ".")
		if len(kk) < 1 {
			formValues[key] = value
			continue
		}
		idx, err := strconv.Atoi(kk[len(kk)-1])
		if err != nil {
			formValues[key] = value
			continue
		}
		newKey := strings.Join(kk[:len(kk)-1], ".")
		list, ok := formValues[newKey]
		if !ok {
			list = []string{}
		}
		for _, v := range value {
			list = append(list, strings.Join([]string{strconv.Itoa(idx), v}, valueSep))
		}
		formValues[newKey] = list
	}
	for key, val := range formValues {
		sort.Strings(val)
		newVal := make([]string, len(val))
		for i, vv := range val {
			v := strings.Split(vv, valueSep)
			if len(v) == 0 {
				continue
			}

			_, err := strconv.Atoi(v[0])
			if err != nil || len(v) <= 1 {
				newVal[i] = v[0]
				continue
			}
			newVal[i] = v[1]
		}
		formValues[key] = newVal
	}

	translationFormByLanguage := map[string]url.Values{}
	for key, value := range formValues {
		language, field, ok := strings.Cut(key, ".")
		if !ok {
			err := fmt.Errorf("%q is not a valid key for updating language translations, expecting <lang>.<field>", key)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		if _, err := t.tStore.ByLanguage(ctx, language); err != nil {
			err := fmt.Errorf("cannot fin language %q: %w", language, err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		translation, ok := translationFormByLanguage[language]
		if !ok {
			translation = make(url.Values)
		}
		translation[field] = value
		translationFormByLanguage[language] = translation
	}

	translations := map[string]*model.Translation{}
	for language, translationForm := range translationFormByLanguage {
		var t model.Translation
		if err := form.Unmarshal(translationForm, &t); err != nil {
			err := fmt.Errorf("unmarshal translation form for language %q: %w", language, err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		translations[language] = &t
	}

	if err := t.tStore.UpdateLanguages(ctx, translations); err != nil {
		err := fmt.Errorf("update languages in store: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
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

func evalTemplateUnsafe(msg string, data any) (string, error) {
	// NOTE: workaround to allow html formatting
	t, err := txttemplate.New("tmp").Parse(msg)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
