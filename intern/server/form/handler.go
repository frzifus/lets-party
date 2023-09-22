package form

import (
	"embed"
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/frzifus/lets-party/intern/db"
	"github.com/frzifus/lets-party/intern/model"
)

//go:embed *.html
var form embed.FS

func NewProcessor(iStore db.InvitationStore, tStore db.TranslationStore, gStore db.GuestStore) *Processor {
	coreTemplates := []string{ "main.html", "footer.html" }
	formTemplates := []string{
		"guest-form.header.html",
		"greeting.html",
		"location.html",
		"date.html",
		"guest-form.html",
	}
	languageTemplates := []string{ "language.html", "language.header.html" }

	return &Processor{
		tmplForm: template.Must(template.ParseFS(form, append(coreTemplates, formTemplates...)...)),
		tmplLang: template.Must(template.ParseFS(form, append(coreTemplates, languageTemplates...)...)),
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
		Date:					time.Date(2023, 12, 24, 0, 0, 0, 0, time.Local),
	}

	p.tmplForm.Execute(c.Writer, gin.H{
		"id":          id,
		"meta":        meta,
		"translation": translation,
		"guests":      guests,
	})
}

func (p *Processor) RenderGuestInputBlock(c *gin.Context, id uuid.UUID) {
	// TODO: get current language from query parameter
	translation, err := p.tStore.ByLanguage(c, "en")
	if err != nil {
		panic(err)
	}

	/**
	 * TODO: Some options
	 *	- remove the wrapperTemplate, directly render guest-input and remove the define from guest-put
	 *	- make it possible to use the guest-input within the guest-form inside the guest-loop
	 *		- this currently fails because without https://gohugo.io/functions/dict/ it seems it is not possible to pass both the $root data and the $guest data (".") to the template
	 *		- missing $.translation data
	 *		- https://stackoverflow.com/questions/18276173/calling-a-template-with-several-pipeline-parameters
	 */
	wrapperTemplate, _ := template.New("wrapper").Parse("{{ block \"GUEST_INPUT\" .}} {{ end }}")
	template.Must(wrapperTemplate.ParseFS(form, "guest-input.html")).Execute(c.Writer, gin.H{
		"ID":          id,
		"translation": translation,
	})
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
