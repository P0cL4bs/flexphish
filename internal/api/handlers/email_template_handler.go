package handlers

import (
	"encoding/json"
	"flexphish/internal/domain/template"
	"net/http"
	"strings"
)

type EmailTemplateHandler struct {
	repo template.EmailTemplateRepository
}

func NewEmailTemplateHandler(repo template.EmailTemplateRepository) *EmailTemplateHandler {
	return &EmailTemplateHandler{repo: repo}
}

func (h *EmailTemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	var input struct {
		Name     string `json:"name"`
		IsGlobal bool   `json:"is_global"`
		Subject  string `json:"subject"`
		Body     string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
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
		Name:     input.Name,
		IsGlobal: input.IsGlobal,
		Subject:  input.Subject,
		Body:     input.Body,
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
		Name     string `json:"name"`
		IsGlobal bool   `json:"is_global"`
		Subject  string `json:"subject"`
		Body     string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
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
	existing.IsGlobal = input.IsGlobal
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
