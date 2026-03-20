package handlers

import (
	"encoding/json"
	"flexphish/internal/domain/smtp"
	"net/http"
	"strings"
)

type SMTPHandler struct {
	repo smtp.Repository
}

func NewSMTPHandler(repo smtp.Repository) *SMTPHandler {
	return &SMTPHandler{repo: repo}
}

func (h *SMTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	var input struct {
		Name      string `json:"name"`
		IsGlobal  bool   `json:"is_global"`
		Host      string `json:"host"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FromName  string `json:"from_name"`
		FromEmail string `json:"from_email"`
		IsActive  bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Host = strings.TrimSpace(input.Host)
	input.Username = strings.TrimSpace(input.Username)
	input.Password = strings.TrimSpace(input.Password)
	input.FromName = strings.TrimSpace(input.FromName)
	input.FromEmail = strings.TrimSpace(input.FromEmail)

	if input.Name == "" || input.Host == "" || input.Port <= 0 || input.Username == "" || input.Password == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "missing required fields",
		})
		return
	}

	connectionExists, err := h.repo.ExistsByConnection(input.Host, input.Port, input.Username, userID, input.IsGlobal, nil)
	if err != nil {
		http.Error(w, "error validating smtp profile", http.StatusInternalServerError)
		return
	}
	if connectionExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": smtp.ErrConnectionAlreadyExists.Error(),
		})
		return
	}

	profile := smtp.SMTPProfile{
		Name:      input.Name,
		IsGlobal:  input.IsGlobal,
		Host:      input.Host,
		Port:      input.Port,
		Username:  input.Username,
		Password:  input.Password,
		FromName:  input.FromName,
		FromEmail: input.FromEmail,
		IsActive:  input.IsActive,
	}

	if !profile.IsGlobal {
		profile.UserId = &userID
	}

	if err := h.repo.Create(&profile); err != nil {
		http.Error(w, "error creating smtp profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

func (h *SMTPHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	profiles, err := h.repo.GetAll(userID)
	if err != nil {
		http.Error(w, "error fetching smtp profiles", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(profiles)
}

func (h *SMTPHandler) Get(w http.ResponseWriter, r *http.Request) {
	profile := r.Context().Value("smtpProfile").(*smtp.SMTPProfile)
	json.NewEncoder(w).Encode(profile)
}

func (h *SMTPHandler) Update(w http.ResponseWriter, r *http.Request) {
	existing := r.Context().Value("smtpProfile").(*smtp.SMTPProfile)
	userID := r.Context().Value("userID").(int64)

	var input struct {
		Name      string `json:"name"`
		IsGlobal  bool   `json:"is_global"`
		Host      string `json:"host"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FromName  string `json:"from_name"`
		FromEmail string `json:"from_email"`
		IsActive  bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Host = strings.TrimSpace(input.Host)
	input.Username = strings.TrimSpace(input.Username)
	input.Password = strings.TrimSpace(input.Password)
	input.FromName = strings.TrimSpace(input.FromName)
	input.FromEmail = strings.TrimSpace(input.FromEmail)

	if input.Name == "" || input.Host == "" || input.Port <= 0 || input.Username == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "missing required fields",
		})
		return
	}

	connectionExists, err := h.repo.ExistsByConnection(input.Host, input.Port, input.Username, userID, input.IsGlobal, &existing.Id)
	if err != nil {
		http.Error(w, "error validating smtp profile", http.StatusInternalServerError)
		return
	}
	if connectionExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": smtp.ErrConnectionAlreadyExists.Error(),
		})
		return
	}

	existing.Name = input.Name
	existing.IsGlobal = input.IsGlobal
	existing.Host = input.Host
	existing.Port = input.Port
	existing.Username = input.Username
	if input.Password != "" {
		existing.Password = input.Password
	}
	existing.FromName = input.FromName
	existing.FromEmail = input.FromEmail
	existing.IsActive = input.IsActive

	if existing.IsGlobal {
		existing.UserId = nil
	} else {
		existing.UserId = &userID
	}

	if err := h.repo.Update(existing); err != nil {
		http.Error(w, "error updating smtp profile", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existing)
}

func (h *SMTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	profile := r.Context().Value("smtpProfile").(*smtp.SMTPProfile)

	if err := h.repo.Delete(profile.Id); err != nil {
		http.Error(w, "error deleting smtp profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
