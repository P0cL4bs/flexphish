package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flexphish/internal/config"
	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/group"
	"flexphish/internal/domain/smtp"
	"flexphish/internal/domain/target"
	"flexphish/internal/domain/template"
	"flexphish/pkg/logger"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

	updatedCampaign, err := h.repo.GetByID(id, userID)
	if err != nil {
		json.NewEncoder(w).Encode(existing)
		return
	}

	json.NewEncoder(w).Encode(updatedCampaign)
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

	if camp.DevMode && camp.SendEmails {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "email sending is not allowed while dev_mode is enabled",
		})
		return
	}

	if camp.SendEmails {
		camp.EmailDispatchStatus = campaign.EmailDispatchQueued
		camp.EmailDispatchQueuedAt = &now
		camp.EmailDispatchStartedAt = nil
		camp.EmailDispatchCompletedAt = nil
		camp.EmailDispatchLastError = ""
		camp.EmailDispatchLastAttemptAt = nil

		if camp.SMTPProfileId == nil || camp.EmailTemplateId == nil {
			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": "smtp_profile_id and email_template_id are required when send_emails is enabled",
			})
			return
		}

		_, err := h.smtpRepo.GetByID(*camp.SMTPProfileId)
		if err != nil {
			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": "smtp profile not found",
			})
			return
		}

		_, err = h.emailTemplateRepo.GetByID(*camp.EmailTemplateId)
		if err != nil {
			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": "email template not found",
			})
			return
		}
	} else {
		camp.EmailDispatchStatus = campaign.EmailDispatchIdle
		camp.EmailDispatchQueuedAt = nil
		camp.EmailDispatchStartedAt = nil
		camp.EmailDispatchCompletedAt = nil
		camp.EmailDispatchLastError = ""
		camp.EmailDispatchLastAttemptAt = nil
		camp.EmailDispatchTotalTargets = 0
		camp.EmailDispatchSent = 0
		camp.EmailDispatchFailed = 0
		camp.EmailDispatchPending = 0
	}

	if err := h.repo.Update(camp); err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	updatedCampaign, err := h.repo.GetByID(id, userID)
	if err != nil {
		JSONResponse(w, http.StatusOK, camp)
		return
	}

	JSONResponse(w, http.StatusOK, updatedCampaign)
}

func (h *CampaignHandler) sendCampaignEmailsInBackground(campaignID int64, userID int64) {
	camp, err := h.repo.GetByID(campaignID, userID)
	if err != nil {
		logger.Log.Error("campaign email dispatch failed to load campaign",
			zap.Int64("campaign_id", campaignID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return
	}

	if !camp.SendEmails || camp.SMTPProfileId == nil || camp.EmailTemplateId == nil {
		logger.Log.Warn("campaign email dispatch skipped due to missing email configuration",
			zap.Int64("campaign_id", campaignID),
			zap.Int64("user_id", userID),
			zap.Bool("send_emails", camp.SendEmails),
		)
		return
	}

	profile, err := h.smtpRepo.GetByID(*camp.SMTPProfileId)
	if err != nil {
		logger.Log.Error("campaign email dispatch failed to load smtp profile",
			zap.Int64("campaign_id", campaignID),
			zap.Int64("smtp_profile_id", *camp.SMTPProfileId),
			zap.Error(err),
		)
		return
	}

	emailTemplate, err := h.emailTemplateRepo.GetByID(*camp.EmailTemplateId)
	if err != nil {
		logger.Log.Error("campaign email dispatch failed to load email template",
			zap.Int64("campaign_id", campaignID),
			zap.Int64("email_template_id", *camp.EmailTemplateId),
			zap.Error(err),
		)
		return
	}

	attachments, err := h.emailTemplateRepo.GetAttachments(emailTemplate.Id)
	if err != nil {
		logger.Log.Error("campaign email dispatch failed to load email template attachments",
			zap.Int64("campaign_id", campaignID),
			zap.Int64("email_template_id", *camp.EmailTemplateId),
			zap.Error(err),
		)
		return
	}

	targets, err := h.collectCampaignTargets(camp.Groups)
	if err != nil || len(targets) == 0 {
		logger.Log.Warn("campaign email dispatch has no eligible targets",
			zap.Int64("campaign_id", campaignID),
			zap.Int("groups_count", len(camp.Groups)),
			zap.Error(err),
		)
		now := time.Now()
		camp.EmailDispatchStatus = campaign.EmailDispatchFailed
		camp.EmailDispatchLastError = "campaign has no eligible targets in selected groups"
		camp.EmailDispatchCompletedAt = &now
		camp.EmailDispatchLastAttemptAt = &now
		camp.EmailDispatchTotalTargets = 0
		camp.EmailDispatchSent = 0
		camp.EmailDispatchFailed = 0
		camp.EmailDispatchPending = 0
		_ = h.repo.Update(camp)
		return
	}

	now := time.Now()
	camp.EmailDispatchStatus = campaign.EmailDispatchProcessing
	camp.EmailDispatchStartedAt = &now
	camp.EmailDispatchCompletedAt = nil
	camp.EmailDispatchLastError = ""
	_ = h.repo.Update(camp)

	totalTargets := int64(len(targets))
	baseURL := h.buildCampaignURL(camp.Subdomain)
	sentCount := int64(0)
	failedCount := int64(0)
	emailsPerMinute := config.GetInt("email_scheduler.emails_per_minute")
	if emailsPerMinute <= 0 {
		emailsPerMinute = 60
	}
	batchSize := config.GetInt("email_scheduler.batch_size")
	if batchSize <= 0 {
		batchSize = 25
	}
	batchPauseMS := config.GetInt("email_scheduler.batch_pause_ms")
	if batchPauseMS < 0 {
		batchPauseMS = 0
	}

	perEmailDelay := time.Minute / time.Duration(emailsPerMinute)
	var lastSendAt time.Time
	processedInBatch := 0

	for _, t := range targets {
		existingTarget, err := h.repo.GetCampaignTargetByTargetID(camp.Id, t.Id)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Log.Warn("campaign email dispatch skipped target due to lookup error",
				zap.Int64("campaign_id", camp.Id),
				zap.Int64("target_id", t.Id),
				zap.Error(err),
			)
			continue
		}

		if existingTarget != nil && existingTarget.Status == "sent" && existingTarget.EmailSentAt != nil {
			sentCount++
			continue
		}

		token := ""
		campaignTarget := &campaign.CampaignTarget{
			CampaignId: camp.Id,
			TargetId:   t.Id,
			Status:     "pending",
		}

		if existingTarget != nil {
			campaignTarget = existingTarget
			token = existingTarget.Token
		}

		if token == "" {
			token, err = generateCampaignToken()
			if err != nil {
				logger.Log.Warn("campaign email dispatch failed to generate token",
					zap.Int64("campaign_id", camp.Id),
					zap.Int64("target_id", t.Id),
					zap.Error(err),
				)
				continue
			}
			campaignTarget.Token = token
		}

		_ = h.repo.SaveCampaignTarget(campaignTarget)

		trackedURL := fmt.Sprintf("%s?s=%s", baseURL, token)
		subject := renderEmailTemplateField(emailTemplate.Subject, t, trackedURL)
		body := renderEmailTemplateField(emailTemplate.Body, t, trackedURL)
		if emailTemplate.TrackOpens {
			pixelURL := fmt.Sprintf("%s/o.gif?s=%s", baseURL, token)
			body = injectOpenTrackingPixel(body, pixelURL)
		}
		msg, fromEmail := buildCampaignEmailMessage(profile, t.Email, subject, body, attachments)

		if !lastSendAt.IsZero() {
			elapsed := time.Since(lastSendAt)
			if elapsed < perEmailDelay {
				time.Sleep(perEmailDelay - elapsed)
			}
		}

		sendErr := sendSMTPMessage(
			profile.Host,
			profile.Port,
			profile.Username,
			profile.Password,
			fromEmail,
			[]string{t.Email},
			msg,
		)
		if sendErr != nil {
			campaignTarget.Status = "failed"
			camp.EmailDispatchLastError = sendErr.Error()
			_ = h.repo.SaveCampaignTarget(campaignTarget)
			logger.Log.Warn("campaign email dispatch failed to send",
				zap.Int64("campaign_id", camp.Id),
				zap.Int64("target_id", t.Id),
				zap.String("email", t.Email),
				zap.Error(sendErr),
			)
			lastSendAt = time.Now()
			processedInBatch++
			failedCount++
			if processedInBatch >= batchSize {
				processedInBatch = 0
				if batchPauseMS > 0 {
					time.Sleep(time.Duration(batchPauseMS) * time.Millisecond)
				}
			}
			continue
		}

		sentAt := time.Now()
		campaignTarget.Status = "sent"
		campaignTarget.EmailSentAt = &sentAt

		if err := h.repo.SaveCampaignTarget(campaignTarget); err != nil {
			lastSendAt = time.Now()
			processedInBatch++
			if processedInBatch >= batchSize {
				processedInBatch = 0
				if batchPauseMS > 0 {
					time.Sleep(time.Duration(batchPauseMS) * time.Millisecond)
				}
			}
			continue
		}

		logger.Log.Info("campaign email sent",
			zap.Int64("campaign_id", camp.Id),
			zap.Int64("target_id", t.Id),
			zap.String("email", t.Email),
		)

		sentCount++
		lastSendAt = time.Now()
		processedInBatch++
		if processedInBatch >= batchSize {
			processedInBatch = 0
			if batchPauseMS > 0 {
				time.Sleep(time.Duration(batchPauseMS) * time.Millisecond)
			}
		}
	}

	camp.TotalSent = sentCount
	camp.EmailDispatchTotalTargets = totalTargets
	camp.EmailDispatchSent = sentCount
	camp.EmailDispatchFailed = failedCount
	camp.EmailDispatchPending = maxInt64(totalTargets-sentCount-failedCount, 0)
	finishedAt := time.Now()
	camp.EmailDispatchCompletedAt = &finishedAt
	camp.EmailDispatchLastAttemptAt = &finishedAt
	if failedCount > 0 || camp.EmailDispatchPending > 0 {
		camp.EmailDispatchStatus = campaign.EmailDispatchFailed
	} else {
		camp.EmailDispatchStatus = campaign.EmailDispatchCompleted
	}
	_ = h.repo.Update(camp)

}

func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func (h *CampaignHandler) collectCampaignTargets(groups []group.Group) ([]target.Target, error) {
	seenByID := make(map[int64]struct{})
	seenByEmail := make(map[string]struct{})
	targets := make([]target.Target, 0)

	for _, g := range groups {
		groupTargets, err := h.groupRepo.ListTargets(g.Id)
		if err != nil {
			return nil, err
		}

		for _, t := range groupTargets {
			if _, exists := seenByID[t.Id]; exists {
				continue
			}

			emailKey := strings.ToLower(strings.TrimSpace(t.Email))
			if emailKey == "" {
				continue
			}

			if _, exists := seenByEmail[emailKey]; exists {
				continue
			}

			seenByID[t.Id] = struct{}{}
			seenByEmail[emailKey] = struct{}{}
			targets = append(targets, t)
		}
	}

	return targets, nil
}

func (h *CampaignHandler) buildCampaignURL(subdomain string) string {
	baseDomain := config.GetString("campaign.base_domain")
	return fmt.Sprintf("http://%s.%s", subdomain, baseDomain)
}

func generateCampaignToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func renderEmailTemplateField(content string, t target.Target, url string) string {
	replacer := strings.NewReplacer(
		"{{FirstName}}", t.FirstName,
		"{{LastName}}", t.LastName,
		"{{Email}}", t.Email,
		"{{Position}}", t.Position,
		"{{URL}}", url,
	)
	return replacer.Replace(content)
}

func injectOpenTrackingPixel(body string, pixelURL string) string {
	pixelTag := fmt.Sprintf(
		`<img src="%s" alt="" width="1" height="1" style="display:none;max-width:1px;max-height:1px;" />`,
		pixelURL,
	)

	lower := strings.ToLower(body)
	bodyEnd := strings.LastIndex(lower, "</body>")
	if bodyEnd >= 0 {
		return body[:bodyEnd] + pixelTag + body[bodyEnd:]
	}
	return body + pixelTag
}

func buildCampaignEmailMessage(
	profile *smtp.SMTPProfile,
	recipient string,
	subject string,
	body string,
	attachments []template.EmailTemplateAttachment,
) ([]byte, string) {
	fromEmail := profile.FromEmail
	if strings.TrimSpace(fromEmail) == "" {
		fromEmail = profile.Username
	}

	fromHeader := fromEmail
	if strings.TrimSpace(profile.FromName) != "" {
		fromHeader = fmt.Sprintf("%s <%s>", profile.FromName, fromEmail)
	}

	boundary := fmt.Sprintf("fph-%d", time.Now().UnixNano())
	var msg bytes.Buffer

	msg.WriteString("From: " + fromHeader + "\r\n")
	msg.WriteString("To: " + recipient + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")

	if len(attachments) == 0 {
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(body + "\r\n")
		return msg.Bytes(), fromEmail
	}

	msg.WriteString("Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body + "\r\n")

	for _, attachment := range attachments {
		safeName := sanitizeMIMEFilename(attachment.Filename)
		mimeType := strings.TrimSpace(attachment.MimeType)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		msg.WriteString("--" + boundary + "\r\n")
		msg.WriteString("Content-Type: " + mimeType + "; name=\"" + safeName + "\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: base64\r\n")
		msg.WriteString("Content-Disposition: attachment; filename=\"" + safeName + "\"\r\n")
		msg.WriteString("\r\n")
		writeBase64WithCRLF(&msg, attachment.Content)
		msg.WriteString("\r\n")
	}

	msg.WriteString("--" + boundary + "--\r\n")
	return msg.Bytes(), fromEmail
}

func sanitizeMIMEFilename(filename string) string {
	name := strings.TrimSpace(filename)
	if name == "" {
		return "attachment.bin"
	}
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "\r", "_")
	name = strings.ReplaceAll(name, "\n", "_")
	return name
}

func writeBase64WithCRLF(buf *bytes.Buffer, content []byte) {
	encoded := base64.StdEncoding.EncodeToString(content)
	const lineLen = 76
	for i := 0; i < len(encoded); i += lineLen {
		end := i + lineLen
		if end > len(encoded) {
			end = len(encoded)
		}
		buf.WriteString(encoded[i:end] + "\r\n")
	}
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
