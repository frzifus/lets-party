package form

import (
	"embed"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/frzifus/lets-party/intern/db"
	"github.com/frzifus/lets-party/intern/model"
)

//go:embed *.html
var form embed.FS

func NewProcessor(iStore db.InvitationStore, tStore db.TranslationStore, gStore db.GuestStore) *Processor {
	return &Processor{
		tmplForm: template.Must(template.ParseFS(form, "form.html")),
		tmplLang: template.Must(template.ParseFS(form, "language.html")),
		iStore:   iStore,
		gStore:   gStore,
		tStore:   tStore,
	}
}

type Processor struct {
	tmplForm *template.Template
	tmplLang *template.Template
	iStore   db.InvitationStore
	gStore   db.GuestStore
	tStore   db.TranslationStore
}

func (p *Processor) Render(c *gin.Context) {
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
		if err != nil {
			c.Error(err)
			return
		}
		if err := p.tmplLang.Execute(c.Writer, gin.H{"id": id, "languages": langs}); err != nil {
			c.Error(err)
		}
		return
	}

	translation, err := p.tStore.ByLanguage(c, lang)
	if err != nil {
		panic(err)
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

	meta := &Metadata{
		Location: Location{
			ZipCode:      "1337",
			Street:       "Milky Way",
			StreetNumber: "42",
			City:         "Somewhere",
			Country:      "Germany",
			Longitudes:   106.6333,
			Latitudes:    10.8167,
		},
	}

	if err := p.tmplForm.Execute(c.Writer, gin.H{
		"id":          id,
		"meta":        meta,
		"translation": translation,
		"guests":      guests,
	}); err != nil {
		panic("ups")
	}
}

type Metadata struct {
	Location Location
	Date     time.Time
}

type Location struct {
	Country      string
	City         string
	ZipCode      string
	Street       string
	StreetNumber string
	Longitudes   float64
	Latitudes    float64
}
