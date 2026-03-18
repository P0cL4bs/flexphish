package controller

import (
	"net/http"
	"strings"

	"flexphish/internal/application/runtime"
	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/template"
)

type CampaignRuntimeController struct {
	campaignRepo   campaign.Repository
	templateRepo   template.TemplateRepository
	sessionService runtime.SessionService
	eventService   runtime.EventService
	stateStore     runtime.StepStateStore
}

func NewCampaignRuntimeController(
	campaignRepo campaign.Repository,
	templateRepo template.TemplateRepository,
	ss runtime.SessionService,
	es runtime.EventService,
	s runtime.StepStateStore,
) *CampaignRuntimeController {

	return &CampaignRuntimeController{
		campaignRepo:   campaignRepo,
		templateRepo:   templateRepo,
		sessionService: ss,
		eventService:   es,
		stateStore:     s,
	}
}

func (c *CampaignRuntimeController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", c.handleRequest)
}

func (c *CampaignRuntimeController) handleRequest(w http.ResponseWriter, r *http.Request) {

	subdomain := extractSubdomain(r.Host)
	if subdomain == "" {
		http.NotFound(w, r)
		return
	}

	camp, err := c.campaignRepo.FindActiveBySubdomain(subdomain)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tmpl, err := c.templateRepo.GetTemplateByFilename(camp.TemplateId)
	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	c.serveTemplate(w, r, camp, tmpl)
}

func extractSubdomain(host string) string {
	host = strings.Split(host, ":")[0]
	parts := strings.Split(host, ".")

	if len(parts) < 2 {
		return ""
	}

	return parts[0]
}

func (c *CampaignRuntimeController) serveTemplate(
	w http.ResponseWriter,
	r *http.Request,
	camp *campaign.Campaign,
	tmpl *template.Template,
) {

	engine := runtime.NewTemplateEngine(
		tmpl,
		camp,
		c.sessionService,
		c.eventService,
		c.stateStore,
	)

	engine.Handler().ServeHTTP(w, r)
}
