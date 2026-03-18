package handlers

import (
	"encoding/base64"
	"encoding/json"
	"flexphish/internal/domain/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type deleteStaticFileRequest struct {
	TemplateFilename string `json:"t_filename"`
	Filename         string `json:"filename"`
}
type createStaticFileRequest struct {
	TemplateFilename string `json:"t_filename"`
	Filename         string `json:"filename"`
	Content          string `json:"content"`
}

type StaticFileHandler struct {
	repo   template.StaticFileRepository
	rtempl template.TemplateRepository
}

func NewStaticFileHandler(repo template.StaticFileRepository, rtempl template.TemplateRepository) *StaticFileHandler {
	return &StaticFileHandler{repo: repo, rtempl: rtempl}
}

func (h *StaticFileHandler) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	if filename == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename required",
		})
		return
	}

	files, err := h.repo.GetAllByTemplateFilename(filename)
	if err != nil {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, files)
}

func (h *StaticFileHandler) Create(w http.ResponseWriter, r *http.Request) {

	var req createRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	if req.TemplateFilename == "" || req.Filename == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "t_filename and filename required",
		})
		return
	}

	if strings.Contains(req.Filename, "..") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid filename",
		})
		return
	}

	templ, _ := h.rtempl.GetTemplateByFilename(req.TemplateFilename)

	if templ.Info.System {
		JSONResponse(w, http.StatusForbidden, map[string]string{
			"error": "system templates cannot be modified",
		})
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid base64 content",
		})
		return
	}

	file, err := h.repo.CreateByTemplateFilename(
		req.TemplateFilename,
		req.Filename,
		data,
	)

	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusCreated, file)
}

func (h *StaticFileHandler) Update(w http.ResponseWriter, r *http.Request) {

	var req createStaticFileRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	if req.TemplateFilename == "" || req.Filename == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "t_filename and filename required",
		})
		return
	}

	if strings.Contains(req.Filename, "..") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid filename",
		})
		return
	}

	templ, _ := h.rtempl.GetTemplateByFilename(req.TemplateFilename)

	if templ.Info.System {
		JSONResponse(w, http.StatusForbidden, map[string]string{
			"error": "system templates cannot be modified",
		})
		return
	}

	file, err := h.repo.UpdateByTemplateFilename(
		req.TemplateFilename,
		req.Filename,
		req.Content,
	)

	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, file)
}

func (h *StaticFileHandler) Delete(w http.ResponseWriter, r *http.Request) {

	var req deleteStaticFileRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	if req.TemplateFilename == "" || req.Filename == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "t_filename and filename required",
		})
		return
	}

	if strings.Contains(req.Filename, "..") {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid filename",
		})
		return
	}
	templ, _ := h.rtempl.GetTemplateByFilename(req.TemplateFilename)

	if templ.Info.System {
		JSONResponse(w, http.StatusForbidden, map[string]string{
			"error": "system templates cannot be modified",
		})
		return
	}

	file, err := h.repo.DeleteByTemplateFilename(
		req.TemplateFilename,
		req.Filename,
	)

	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, file)
}
