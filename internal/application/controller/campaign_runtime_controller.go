package controller

import (
	"net/http"
	"strings"
	"time"

	"flexphish/internal/application/runtime"
	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/event"
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

	if r.URL.Path == "/o.gif" {
		c.handleOpenTrackingPixel(w, r, camp)
		return
	}

	tmpl, err := c.templateRepo.GetTemplateByFilename(camp.TemplateId)
	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	c.serveTemplate(w, r, camp, tmpl)
}

func (c *CampaignRuntimeController) handleOpenTrackingPixel(w http.ResponseWriter, r *http.Request, camp *campaign.Campaign) {
	token := strings.TrimSpace(r.URL.Query().Get("s"))
	if token != "" {
		campaignTarget, err := c.campaignRepo.GetCampaignTargetByToken(camp.Id, token)
		if err == nil && campaignTarget != nil {
			info := runtime.ExtractVisitorInfo(r)
			openedNow, markErr := c.campaignRepo.MarkCampaignTargetOpenedIfFirst(
				campaignTarget.Id,
				nil,
				info.IP,
				info.UserAgent,
				time.Now(),
			)
			if markErr == nil && openedNow {
				_ = c.campaignRepo.IncrementOpened(camp.Id)
			}

			_ = c.eventService.RegisterEvent(
				camp.Id,
				campaignTarget.ResultId,
				event.EventOpen,
				"email_open_pixel",
				r,
				map[string]interface{}{
					"source": "email_pixel",
				},
			)
		}
	}

	// 1x1 transparent GIF
	pixel := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00,
		0x01, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xff, 0xff, 0xff, 0x21, 0xf9, 0x04, 0x01, 0x00,
		0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
		0x01, 0x00, 0x3b,
	}
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pixel)
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
