package handlers

import (
	"encoding/json"
	"errors"
	"flexphish/internal/domain/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type EmailTemplateHandler struct {
	repo template.EmailTemplateRepository
}

const maxEmailTemplateAttachmentSize = 10 * 1024 * 1024

func NewEmailTemplateHandler(repo template.EmailTemplateRepository) *EmailTemplateHandler {
	return &EmailTemplateHandler{repo: repo}
}

func (h *EmailTemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	var input struct {
		Name       string `json:"name"`
		Category   string `json:"category"`
		IsGlobal   bool   `json:"is_global"`
		TrackOpens *bool  `json:"track_opens"`
		Subject    string `json:"subject"`
		Body       string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Category = strings.TrimSpace(input.Category)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Body = strings.TrimSpace(input.Body)

	if input.Name == "" || input.Subject == "" || input.Body == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "name, subject and body are required",
		})
		return
	}

	nameExists, err := h.repo.ExistsByName(input.Name, userID, input.IsGlobal, nil)
	if err != nil {
		http.Error(w, "error validating email template", http.StatusInternalServerError)
		return
	}
	if nameExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "email template name already exists",
		})
		return
	}

	emailTemplate := template.EmailTemplate{
		Name:       input.Name,
		Category:   input.Category,
		IsGlobal:   input.IsGlobal,
		TrackOpens: true,
		Subject:    input.Subject,
		Body:       input.Body,
	}

	if input.TrackOpens != nil {
		emailTemplate.TrackOpens = *input.TrackOpens
	}

	if !emailTemplate.IsGlobal {
		emailTemplate.UserId = &userID
	}

	if err := h.repo.Create(&emailTemplate); err != nil {
		http.Error(w, "error creating email template", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(emailTemplate)
}

func (h *EmailTemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	emailTemplates, err := h.repo.GetAll(userID)
	if err != nil {
		http.Error(w, "error fetching email templates", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(emailTemplates)
}

func (h *EmailTemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	emailTemplate := r.Context().Value("emailTemplate").(*template.EmailTemplate)
	json.NewEncoder(w).Encode(emailTemplate)
}

func (h *EmailTemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	existing := r.Context().Value("emailTemplate").(*template.EmailTemplate)
	userID := r.Context().Value("userID").(int64)

	var input struct {
		Name       string `json:"name"`
		Category   string `json:"category"`
		IsGlobal   bool   `json:"is_global"`
		TrackOpens *bool  `json:"track_opens"`
		Subject    string `json:"subject"`
		Body       string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Category = strings.TrimSpace(input.Category)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Body = strings.TrimSpace(input.Body)

	if input.Name == "" || input.Subject == "" || input.Body == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "name, subject and body are required",
		})
		return
	}

	nameExists, err := h.repo.ExistsByName(input.Name, userID, input.IsGlobal, &existing.Id)
	if err != nil {
		http.Error(w, "error validating email template", http.StatusInternalServerError)
		return
	}
	if nameExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "email template name already exists",
		})
		return
	}

	existing.Name = input.Name
	existing.Category = input.Category
	existing.IsGlobal = input.IsGlobal
	if input.TrackOpens != nil {
		existing.TrackOpens = *input.TrackOpens
	}
	existing.Subject = input.Subject
	existing.Body = input.Body

	if existing.IsGlobal {
		existing.UserId = nil
	} else {
		existing.UserId = &userID
	}

	if err := h.repo.Update(existing); err != nil {
		http.Error(w, "error updating email template", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existing)
}

func (h *EmailTemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	emailTemplate := r.Context().Value("emailTemplate").(*template.EmailTemplate)

	if err := h.repo.Delete(emailTemplate.Id); err != nil {
		http.Error(w, "error deleting email template", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EmailTemplateHandler) ListAttachments(w http.ResponseWriter, r *http.Request) {
	emailTemplate := r.Context().Value("emailTemplate").(*template.EmailTemplate)

	attachments, err := h.repo.GetAttachments(emailTemplate.Id)
	if err != nil {
		http.Error(w, "error fetching attachments", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(attachments)
}

func (h *EmailTemplateHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	emailTemplate := r.Context().Value("emailTemplate").(*template.EmailTemplate)

	if err := r.ParseMultipartForm(maxEmailTemplateAttachmentSize); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid multipart form",
		})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "file is required",
		})
		return
	}
	defer file.Close()

	filename := strings.TrimSpace(header.Filename)
	if filename == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "filename is required",
		})
		return
	}

	limitedReader := io.LimitReader(file, maxEmailTemplateAttachmentSize+1)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		http.Error(w, "error reading attachment", http.StatusInternalServerError)
		return
	}
	if len(content) == 0 {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "attachment cannot be empty",
		})
		return
	}
	if len(content) > maxEmailTemplateAttachmentSize {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "attachment exceeds 10MB limit",
		})
		return
	}

	mimeType := http.DetectContentType(content)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	attachment := template.EmailTemplateAttachment{
		EmailTemplateId: emailTemplate.Id,
		Filename:        filename,
		MimeType:        mimeType,
		Size:            int64(len(content)),
		Content:         content,
	}

	if err := h.repo.CreateAttachment(&attachment); err != nil {
		http.Error(w, "error creating attachment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(attachment)
}

func (h *EmailTemplateHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	emailTemplate := r.Context().Value("emailTemplate").(*template.EmailTemplate)

	vars := mux.Vars(r)
	attachmentID, err := strconv.ParseInt(vars["attachmentId"], 10, 64)
	if err != nil || attachmentID <= 0 {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid attachment id",
		})
		return
	}

	if err := h.repo.DeleteAttachment(emailTemplate.Id, attachmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			JSONResponse(w, http.StatusNotFound, map[string]string{
				"error": "attachment not found",
			})
			return
		}
		http.Error(w, "error deleting attachment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
