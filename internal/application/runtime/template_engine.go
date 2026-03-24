package runtime

import (
	"encoding/json"
	"flexphish/internal/api/handlers"
	"flexphish/internal/config"
	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/event"
	"flexphish/internal/domain/template"
	htmltmpl "html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type TemplateEngine struct {
	tmpl           *template.Template
	campaign       *campaign.Campaign
	sessionService SessionService
	eventService   EventService
	stateStore     StepStateStore
}

type StepResponse struct {
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Redirect string `json:"redirect,omitempty"`
}

func NewTemplateEngine(
	t *template.Template,
	camp *campaign.Campaign,
	ss SessionService,
	es EventService,
	s StepStateStore,
) *TemplateEngine {

	return &TemplateEngine{
		tmpl:           t,
		campaign:       camp,
		sessionService: ss,
		eventService:   es,
		stateStore:     s,
	}
}

func (e *TemplateEngine) Handler() http.Handler {

	mux := http.NewServeMux()

	staticPath := filepath.Join(e.tmpl.TemplateDir, "static")
	fs := http.FileServer(http.Dir(staticPath))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	if len(e.tmpl.Steps) > 0 {
		first := e.tmpl.Steps[0]
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			redirectURL := first.Path
			if r.URL.RawQuery != "" {
				redirectURL += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, redirectURL, http.StatusFound)
		})
	}

	for _, step := range e.tmpl.Steps {
		mux.HandleFunc(step.Path, e.makeStepHandler(step))
	}

	return mux
}

func (e *TemplateEngine) makeStepHandler(step template.Step) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		testToken := r.URL.Query().Get("test_mode_token")
		configToken := config.GetString("security.test_mode_token")
		isTestMode := testToken != "" && testToken == configToken
		e.eventService.SetTestMode(isTestMode)
		e.sessionService.SetTestMode(isTestMode)
		isFinal := step.Next == "" || step.RedirectURL != ""

		campaignId := e.campaign.Id

		campaignToken := strings.TrimSpace(r.URL.Query().Get("s"))
		res, err := e.sessionService.Resolve(w, r, campaignId, campaignToken)
		if err != nil {
			handlers.JSONResponse(w, http.StatusInternalServerError, StepResponse{
				Success: false,
				Error:   "session error",
			})
			return
		}

		if r.Method == http.MethodGet && r.URL.Path == step.Path {
			_ = e.eventService.RegisterEvent(
				campaignId,
				&res.Id,
				event.EventPageView,
				step.ID,
				r,
				map[string]interface{}{
					"path":       r.URL.Path,
					"query":      r.URL.RawQuery,
					"referer":    r.Referer(),
					"user_agent": r.UserAgent(),
				},
			)
		}

		if r.Method == step.Method {

			if step.SimulateDelayMS > 0 {
				time.Sleep(time.Duration(step.SimulateDelayMS) * time.Millisecond)
			}

			if err := r.ParseForm(); err != nil {
				e.eventService.RegisterEvent(
					campaignId,
					&res.Id,
					event.EventError,
					step.ID,
					r,
					map[string]interface{}{
						"error": "form_parse_error",
					},
				)
				handlers.JSONResponse(w, http.StatusBadRequest, StepResponse{
					Success: false,
					Error:   "invalid form",
				})
				return
			}

			captured := map[string]interface{}{}

			for _, field := range step.Capture.Fields {

				val := r.FormValue(field.Name)

				if field.Required && val == "" {

					e.eventService.RegisterEvent(
						campaignId,
						&res.Id,
						event.EventError,
						step.ID,
						r,
						map[string]interface{}{
							"type":  "validation_error",
							"field": field.Name,
						},
					)

					handlers.JSONResponse(w, http.StatusBadRequest, StepResponse{
						Success: false,
						Error:   field.ErrorMessage,
					})
					return
				}

				if field.ValidateRegex != "" {
					matched, _ := regexp.MatchString(field.ValidateRegex, val)
					if !matched {

						e.eventService.RegisterEvent(
							campaignId,
							&res.Id,
							event.EventError,
							step.ID,
							r,
							map[string]interface{}{
								"type":  "regex_validation_failed",
								"field": field.Name,
							},
						)

						handlers.JSONResponse(w, http.StatusBadRequest, StepResponse{
							Success: false,
							Error:   field.ErrorMessage,
						})
						return
					}
				}

				captured[field.Name] = val
			}
			_ = e.stateStore.Merge(res.SessionID, captured)

			_ = e.eventService.RegisterStepSubmit(
				campaignId,
				res,
				step.ID,
				captured,
				r,
				isFinal,
			)

			if step.Next != "" {
				for _, s := range e.tmpl.Steps {
					if s.ID == step.Next {

						e.eventService.RegisterEvent(
							campaignId,
							&res.Id,
							event.EventRedirect,
							step.ID,
							r,
							map[string]interface{}{
								"to": s.Path,
							},
						)

						handlers.JSONResponse(w, http.StatusOK, StepResponse{
							Success:  true,
							Redirect: s.Path,
						})
						return
					}
				}
			}
			if step.RedirectURL != "" {

				e.eventService.RegisterEvent(
					campaignId,
					&res.Id,
					event.EventRedirect,
					step.ID,
					r,
					map[string]interface{}{
						"to": step.RedirectURL,
					},
				)

				w.Header().Set("Content-Type", "application/json")

				handlers.JSONResponse(w, http.StatusOK, StepResponse{
					Success:  true,
					Redirect: step.RedirectURL,
				})

				return
			}
		}

		tmplPath := filepath.Join(
			e.tmpl.TemplateDir,
			step.TemplateFile,
		)

		tpl, err := htmltmpl.New(filepath.Base(tmplPath)).
			Funcs(htmltmpl.FuncMap{
				"marshal": func(v interface{}) htmltmpl.JS {
					b, _ := json.Marshal(v)
					return htmltmpl.JS(b)
				},
			}).
			ParseFiles(tmplPath)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err != nil {
			http.Error(w, "Template render error", http.StatusInternalServerError)
			return
		}
		stateVars := map[string]interface{}{}
		if v, err := e.stateStore.Get(res.SessionID); err == nil && v != nil {
			stateVars = v
		}
		contextVars := map[string]interface{}{
			"campaign_id": campaignId,
			"session_id":  res.Id,
			"ip":          r.RemoteAddr,
			"user_agent":  r.UserAgent(),
			"query":       r.URL.Query(),
		}
		for k, v := range e.tmpl.GlobalVars {
			contextVars[k] = v
		}

		for k, v := range step.Vars {
			contextVars[k] = v
		}

		for k, v := range stateVars {
			contextVars[k] = v
		}

		_ = tpl.Execute(w, map[string]interface{}{
			"Hooks": e.tmpl.Hooks.OnLoad,
			"Vars":  contextVars,
		})
	}
}
