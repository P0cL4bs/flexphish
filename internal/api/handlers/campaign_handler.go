package handlers

import (
	"encoding/json"
	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/group"
	"flexphish/internal/domain/smtp"
	"flexphish/internal/domain/template"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type CampaignHandler struct {
	repo              campaign.Repository
	trepo             template.TemplateRepository
	groupRepo         group.Repository
	smtpRepo          smtp.Repository
	emailTemplateRepo template.EmailTemplateRepository
}

func NewCampaignHandler(
	repo campaign.Repository,
	trepo template.TemplateRepository,
	groupRepo group.Repository,
	smtpRepo smtp.Repository,
	emailTemplateRepo template.EmailTemplateRepository,
) *CampaignHandler {
	return &CampaignHandler{
		repo:              repo,
		trepo:             trepo,
		groupRepo:         groupRepo,
		smtpRepo:          smtpRepo,
		emailTemplateRepo: emailTemplateRepo,
	}
}
func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(int64)
	var input struct {
		Name            string  `json:"name"`
		TemplateID      string  `json:"template_id"`
		Subdomain       string  `json:"subdomain"`
		DevMode         bool    `json:"dev_mode"`
		GroupIDs        []int64 `json:"group_ids"`
		SMTPProfileID   int64   `json:"smtp_profile_id"`
		EmailTemplateID int64   `json:"email_template_id"`
		SendEmails      bool    `json:"send_emails"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.TemplateID == "" || input.Subdomain == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	if exist, _ := h.trepo.Exists(input.TemplateID); !exist {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "template_id does not exist",
		})
		return
	}

	groups, err := h.resolveGroups(userID, input.GroupIDs)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	smtpProfileID, err := h.resolveSMTPProfileID(userID, input.SMTPProfileID)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	emailTemplateID, err := h.resolveEmailTemplateID(userID, input.EmailTemplateID)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	newCampaign := &campaign.Campaign{
		UserId:          userID,
		Name:            input.Name,
		Subdomain:       input.Subdomain,
		TemplateId:      input.TemplateID,
		Status:          campaign.StatusDraft,
		DevMode:         input.DevMode,
		Groups:          groups,
		SMTPProfileId:   smtpProfileID,
		EmailTemplateId: emailTemplateID,
		SendEmails:      input.SendEmails || (smtpProfileID != nil && emailTemplateID != nil),
		SMTPProfile:     nil,
		EmailTemplate:   nil,
		CampaignTargets: nil,
		Results:         nil,
		Events:          nil,
	}
	if existing, err := h.repo.FindBySubdomain(input.Subdomain); err == nil {
		if existing != nil {
			JSONResponse(w, http.StatusConflict, map[string]string{
				"error": "subdomain already exists",
			})
			return
		}
	}

	if err := h.repo.Create(newCampaign); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "could not create campaign",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newCampaign)
}

func (h *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(int64)

	query := r.URL.Query()

	statusParam := query.Get("status")
	pageParam := query.Get("page")
	pageSizeParam := query.Get("page_size")

	page, _ := strconv.Atoi(pageParam)
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(pageSizeParam)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	var status *campaign.CampaignStatus
	if statusParam != "" {
		s := campaign.CampaignStatus(statusParam)
		status = &s
	}

	campaigns, total, err := h.repo.ListByUser(userID, status, page, pageSize)
	if err != nil {
		http.Error(w, "could not fetch campaigns", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":      campaigns,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *CampaignHandler) GetByID(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(int64)

	vars := mux.Vars(r)
	idParam := vars["id"]

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid id",
		})
		return
	}
	camp, err := h.repo.GetByID(id, userID)
	if err != nil {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "campaign not found",
		})
		return
	}

	json.NewEncoder(w).Encode(camp)
}

func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(int64)

	vars := mux.Vars(r)
	idParam := vars["id"]

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	existing, err := h.repo.GetByID(id, userID)
	if err != nil {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	}

	var input struct {
		Name              *string  `json:"name"`
		Status            *string  `json:"status"`
		TemplateRequestId *string  `json:"template_id"`
		DevMode           *bool    `json:"dev_mode"`
		GroupIDs          *[]int64 `json:"group_ids"`
		SMTPProfileID     *int64   `json:"smtp_profile_id"`
		EmailTemplateID   *int64   `json:"email_template_id"`
		SendEmails        *bool    `json:"send_emails"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if input.Name != nil {
		existing.Name = *input.Name
	}

	if input.Status != nil {
		existing.Status = campaign.CampaignStatus(*input.Status)
	}

	if input.TemplateRequestId != nil {
		existing.TemplateId = *input.TemplateRequestId
	}

	if input.DevMode != nil {
		existing.DevMode = *input.DevMode
	}

	if input.GroupIDs != nil {
		groups, err := h.resolveGroups(userID, *input.GroupIDs)
		if err != nil {
			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		existing.Groups = groups
	}

	if input.SMTPProfileID != nil {
		smtpProfileID, err := h.resolveSMTPProfileID(userID, *input.SMTPProfileID)
		if err != nil {
			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}
		existing.SMTPProfileId = smtpProfileID
	}

	if input.EmailTemplateID != nil {
		emailTemplateID, err := h.resolveEmailTemplateID(userID, *input.EmailTemplateID)
		if err != nil {
			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}
		existing.EmailTemplateId = emailTemplateID
	}

	if input.SMTPProfileID != nil || input.EmailTemplateID != nil {
		existing.SendEmails = existing.SMTPProfileId != nil && existing.EmailTemplateId != nil
	}

	if input.SendEmails != nil {
		existing.SendEmails = *input.SendEmails
	}

	if exist, _ := h.trepo.Exists(existing.TemplateId); !exist {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "template_id does not exist",
		})
		return
	}

	if existing.Status == campaign.StatusActive && (existing.TemplateId == "" || existing.Subdomain == "") {
		http.Error(w, "launched campaigns must have template_id and subdomain", http.StatusBadRequest)
		return
	}

	if existing.Status == campaign.StatusActive {
		if existing.LaunchDate == nil {
			now := time.Now()
			existing.LaunchDate = &now
		}
	} else {
		existing.LaunchDate = nil
		existing.CompletedDate = nil
	}

	if existing.Status == campaign.StatusCompleted && existing.CompletedDate == nil {
		now := time.Now()
		existing.CompletedDate = &now
	} else if existing.Status != campaign.StatusCompleted {
		existing.CompletedDate = nil
	}

	if err := h.repo.Update(existing); err != nil {
		http.Error(w, "could not update", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existing)
}

func (h *CampaignHandler) resolveGroups(userID int64, groupIDs []int64) ([]group.Group, error) {
	if len(groupIDs) == 0 {
		return []group.Group{}, nil
	}

	visibleGroups, err := h.groupRepo.GetAll(userID)
	if err != nil {
		return nil, fmt.Errorf("could not validate groups")
	}

	allowed := make(map[int64]struct{}, len(visibleGroups))
	for _, g := range visibleGroups {
		allowed[g.Id] = struct{}{}
	}

	seen := make(map[int64]struct{})
	resolved := make([]group.Group, 0, len(groupIDs))
	for _, id := range groupIDs {
		if id <= 0 {
			continue
		}
		if _, alreadyIncluded := seen[id]; alreadyIncluded {
			continue
		}

		if _, ok := allowed[id]; !ok {
			return nil, fmt.Errorf("group_id %d is invalid", id)
		}

		resolved = append(resolved, group.Group{Id: id})
		seen[id] = struct{}{}
	}

	return resolved, nil
}

func (h *CampaignHandler) resolveSMTPProfileID(userID int64, smtpProfileID int64) (*int64, error) {
	if smtpProfileID <= 0 {
		return nil, nil
	}

	profiles, err := h.smtpRepo.GetAll(userID)
	if err != nil {
		return nil, fmt.Errorf("could not validate smtp_profile_id")
	}

	for _, profile := range profiles {
		if profile.Id == smtpProfileID {
			id := smtpProfileID
			return &id, nil
		}
	}

	return nil, fmt.Errorf("smtp_profile_id %d is invalid", smtpProfileID)
}

func (h *CampaignHandler) resolveEmailTemplateID(userID int64, emailTemplateID int64) (*int64, error) {
	if emailTemplateID <= 0 {
		return nil, nil
	}

	templates, err := h.emailTemplateRepo.GetAll(userID)
	if err != nil {
		return nil, fmt.Errorf("could not validate email_template_id")
	}

	for _, emailTemplate := range templates {
		if emailTemplate.Id == emailTemplateID {
			id := emailTemplateID
			return &id, nil
		}
	}

	return nil, fmt.Errorf("email_template_id %d is invalid", emailTemplateID)
}

func (h *CampaignHandler) Delete(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(int64)

	vars := mux.Vars(r)
	idParam := vars["id"]

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id, userID); err != nil {
		http.Error(w, "could not delete", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "deleted successfully",
	})
}

func (h *CampaignHandler) Start(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(int64)
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid campaign id",
		})
		return
	}

	camp, err := h.repo.GetByID(id, userID)
	if err != nil {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "campaign not found",
		})
		return
	}

	if camp.Status == campaign.StatusActive {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "campaign already active",
		})
		return
	}

	if camp.Status == campaign.StatusCompleted || camp.Status == campaign.StatusCancelled {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "campaign cannot be started",
		})
		return
	}

	now := time.Now()

	camp.Status = campaign.StatusActive
	camp.LaunchDate = &now

	if err := h.repo.Update(camp); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, camp)
}

func (h *CampaignHandler) Stop(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(int64)

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid campaign id",
		})
		return
	}

	camp, err := h.repo.GetByID(id, userID)
	if err != nil {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "campaign not found",
		})
		return
	}

	if camp.Status != campaign.StatusActive {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "campaign is not active",
		})
		return
	}

	camp.Status = campaign.StatusStopped

	if err := h.repo.Update(camp); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, camp)
}

func (h *CampaignHandler) Archive(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(int64)

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid campaign id",
		})
		return
	}

	camp, err := h.repo.GetByID(id, userID)
	if err != nil {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "campaign not found",
		})
		return
	}

	if camp.IsArchived {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": "campaign already archived",
		})
		return
	}

	camp.IsArchived = true

	if err := h.repo.Update(camp); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, camp)
}

func (h *CampaignHandler) Analytics(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(int64)

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}

	analytics, err := h.repo.GetAnalytics(userID, period)
	if err != nil {
		http.Error(w, "could not fetch analytics", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(analytics)
}

func (h *CampaignHandler) DeleteResult(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("userID").(int64)

	vars := mux.Vars(r)

	campaignIDStr := vars["id"]
	resultIDStr := vars["result_id"]

	campaignID, err := strconv.ParseInt(campaignIDStr, 10, 64)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid campaign id",
		})
		return
	}

	resultID, err := strconv.ParseInt(resultIDStr, 10, 64)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid result id",
		})
		return
	}

	_, err = h.repo.GetByID(campaignID, userID)
	if err != nil {
		JSONResponse(w, http.StatusNotFound, map[string]string{
			"error": "campaign not found",
		})
		return
	}

	err = h.repo.DeleteResult(resultID, campaignID)

	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "could not delete result",
		})
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{
		"message": "result deleted successfully",
	})
}
