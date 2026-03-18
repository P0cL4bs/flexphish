package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/template"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

type TemplateHandler struct {
	repo    template.TemplateRepository
	repoCam campaign.Repository
}

type templateRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type deleteTemplateRequest struct {
	Filename string `json:"filename"`
}

func buildTemplateDir(filename string) (string, error) {

	name := strings.TrimSuffix(filename, ".yaml")

	valid := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !valid.MatchString(name) {
		return "", errors.New("invalid filename")
	}

	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, "-", "_")

	return slug, nil
}

func NewTemplateHandler(repo template.TemplateRepository, camrepo campaign.Repository) *TemplateHandler {
	return &TemplateHandler{
		repo:    repo,
		repoCam: camrepo,
	}
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	templates, err := h.repo.LoadAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"templates": templates,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TemplateHandler) GetByFilename(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	template, err := h.repo.GetTemplateByFilename(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func (h *TemplateHandler) GetMetadataByFilename(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	template, err := h.repo.LoadByFilename(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {

	var req templateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	if req.Filename == "" || req.Content == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename and content required",
		})
		return
	}

	if !strings.HasSuffix(req.Filename, ".yaml") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename must have .yaml extension",
		})
		return
	}

	name := strings.TrimSuffix(req.Filename, ".yaml")

	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid filename",
		})
		return
	}

	exists, err := h.repo.Exists(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if exists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "template already exists",
		})
		return
	}

	dirExists, err := h.repo.TemplateDirExists(name)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if dirExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "assets directory already exists",
		})
		return
	}

	var tpl template.Template

	if err := yaml.Unmarshal([]byte(req.Content), &tpl); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid yaml content",
		})
		return
	}
	tpl.Info.System = false
	tpl.TemplateDir, _ = buildTemplateDir(req.Filename)

	updatedYaml, err := yaml.Marshal(&tpl)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.repo.Save(req.Filename, string(updatedYaml)); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.repo.CreateTemplateDir(tpl.TemplateDir); err != nil {
		h.repo.Delete(req.Filename)
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusCreated, tpl)
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {

	var req templateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	if req.Filename == "" || req.Content == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename and content required",
		})
		return
	}

	if !strings.HasSuffix(req.Filename, ".yaml") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid filename",
		})
		return
	}

	exists, err := h.repo.Exists(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if !exists {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "template not found",
		})
		return
	}

	temp, _ := h.repo.GetTemplateByFilename(req.Filename)

	if temp.Info.System {
		JSONResponse(w, http.StatusForbidden, map[string]string{
			"error": "system templates cannot be modified",
		})
		return
	}

	active, err := h.repoCam.HasActiveCampaignUsingTemplate(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to verify template usage",
		})
		return
	}

	if active {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "template cannot be modified because there is an active campaign using it",
		})
		return
	}

	var tpl template.Template

	if err := yaml.Unmarshal([]byte(req.Content), &tpl); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid yaml content",
		})
		return
	}
	tpl.Info.System = false

	expectedDir, err := buildTemplateDir(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	tpl.TemplateDir = expectedDir

	updatedYaml, err := yaml.Marshal(&tpl)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.repo.Save(req.Filename, string(updatedYaml)); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, tpl)
}

func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {

	var req deleteTemplateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	if req.Filename == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename required",
		})
		return
	}

	if !strings.HasSuffix(req.Filename, ".yaml") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid filename",
		})
		return
	}

	exists, err := h.repo.Exists(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if !exists {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "template not found",
		})
		return
	}

	count, err := h.repoCam.CountCampaignsUsingTemplateId(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to check template usage",
		})
		return
	}
	if count > 0 {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "template is used by campaigns",
		})
		return
	}

	data, err := h.repo.LoadByFilename(req.Filename)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	data, _ = h.repo.LoadByFilename(req.Filename)

	templ, _ := h.repo.GetTemplateByFilename(req.Filename)

	if templ.Info.System {
		JSONResponse(w, http.StatusForbidden, map[string]string{
			"error": "system templates cannot be modified",
		})
		return
	}

	if err := h.repo.Delete(req.Filename); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := h.repo.DeleteTemplateDir(data.TemplateDir); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, data)
}
