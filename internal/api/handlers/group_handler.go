package handlers

import (
	"encoding/json"
	"flexphish/internal/domain/group"
	"flexphish/internal/domain/target"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type GroupHandler struct {
	repo group.Repository
}

func NewGroupHandler(repo group.Repository) *GroupHandler {
	return &GroupHandler{repo: repo}
}

func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	var input struct {
		Name     string `json:"name"`
		IsGlobal bool   `json:"is_global"`
		Targets  []struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Position  string `json:"position"`
		} `json:"targets"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "name is required",
		})
		return
	}

	alreadyExists, err := h.repo.ExistsByName(input.Name, userID, input.IsGlobal, nil)
	if err != nil {
		http.Error(w, "error validating group", http.StatusInternalServerError)
		return
	}
	if alreadyExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": group.ErrNameAlreadyExists.Error(),
		})
		return
	}

	targets := make([]target.Target, 0, len(input.Targets))
	for _, t := range input.Targets {
		email := strings.TrimSpace(t.Email)
		if email == "" {
			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": group.ErrInvalidTargets.Error(),
			})
			return
		}

		targets = append(targets, target.Target{
			UserId:    userID,
			FirstName: strings.TrimSpace(t.FirstName),
			LastName:  strings.TrimSpace(t.LastName),
			Email:     email,
			Position:  strings.TrimSpace(t.Position),
		})
	}

	g := group.Group{
		Name:     input.Name,
		IsGlobal: input.IsGlobal,
		Targets:  targets,
	}

	if !g.IsGlobal {
		g.UserId = &userID
	}
	if err := h.repo.Create(&g); err != nil {
		http.Error(w, "error creating group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(g)
}

func (h *GroupHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	groups, err := h.repo.GetAll(userID)
	if err != nil {
		http.Error(w, "error fetching groups", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(groups)
}

func (h *GroupHandler) Get(w http.ResponseWriter, r *http.Request) {
	g := r.Context().Value("group").(*group.Group)
	json.NewEncoder(w).Encode(g)
}

func (h *GroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	existing := r.Context().Value("group").(*group.Group)
	userID := r.Context().Value("userID").(int64)

	var input group.Group
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "name is required",
		})
		return
	}

	alreadyExists, err := h.repo.ExistsByName(input.Name, userID, input.IsGlobal, &existing.Id)
	if err != nil {
		http.Error(w, "error validating group", http.StatusInternalServerError)
		return
	}
	if alreadyExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": group.ErrNameAlreadyExists.Error(),
		})
		return
	}

	existing.Name = input.Name
	existing.IsGlobal = input.IsGlobal
	if existing.IsGlobal {
		existing.UserId = nil
	} else {
		existing.UserId = &userID
	}

	if err := h.repo.Update(existing); err != nil {
		http.Error(w, "error updating group", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existing)
}

func (h *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	g := r.Context().Value("group").(*group.Group)

	if err := h.repo.Delete(g.Id); err != nil {
		http.Error(w, "error deleting group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) ListTargets(w http.ResponseWriter, r *http.Request) {
	g := r.Context().Value("group").(*group.Group)

	targets, err := h.repo.ListTargets(g.Id)
	if err != nil {
		http.Error(w, "error fetching targets", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(targets)
}

func (h *GroupHandler) GetTarget(w http.ResponseWriter, r *http.Request) {
	g := r.Context().Value("group").(*group.Group)

	targetID, err := parseTargetID(r)
	if err != nil {
		http.Error(w, "invalid target id", http.StatusBadRequest)
		return
	}

	t, err := h.repo.GetTargetByID(g.Id, targetID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "target not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error fetching target", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(t)
}

func (h *GroupHandler) CreateTarget(w http.ResponseWriter, r *http.Request) {
	g := r.Context().Value("group").(*group.Group)
	userID := r.Context().Value("userID").(int64)

	input, ok := decodeAndValidateTargetInput(w, r)
	if !ok {
		return
	}

	t := target.Target{
		UserId:    userID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Position:  input.Position,
	}

	emailExists, err := h.repo.TargetEmailExistsInGroup(g.Id, t.Email, nil)
	if err != nil {
		http.Error(w, "error validating target", http.StatusInternalServerError)
		return
	}
	if emailExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": group.ErrTargetEmailExists.Error(),
		})
		return
	}

	if err := h.repo.CreateTarget(g.Id, &t); err != nil {
		http.Error(w, "error creating target", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (h *GroupHandler) UpdateTarget(w http.ResponseWriter, r *http.Request) {
	g := r.Context().Value("group").(*group.Group)

	targetID, err := parseTargetID(r)
	if err != nil {
		http.Error(w, "invalid target id", http.StatusBadRequest)
		return
	}

	input, ok := decodeAndValidateTargetInput(w, r)
	if !ok {
		return
	}

	t := target.Target{
		Id:        targetID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Position:  input.Position,
	}

	if err := h.repo.UpdateTarget(g.Id, &t); err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "target not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error updating target", http.StatusInternalServerError)
		return
	}

	updated, err := h.repo.GetTargetByID(g.Id, targetID)
	if err != nil {
		http.Error(w, "error fetching updated target", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *GroupHandler) DeleteTarget(w http.ResponseWriter, r *http.Request) {
	g := r.Context().Value("group").(*group.Group)

	targetID, err := parseTargetID(r)
	if err != nil {
		http.Error(w, "invalid target id", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteTarget(g.Id, targetID); err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "target not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error deleting target", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type targetInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Position  string `json:"position"`
}

func decodeAndValidateTargetInput(w http.ResponseWriter, r *http.Request) (*targetInput, bool) {
	var input targetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return nil, false
	}

	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Email = strings.TrimSpace(input.Email)
	input.Position = strings.TrimSpace(input.Position)

	if input.Email == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": group.ErrInvalidTargets.Error(),
		})
		return nil, false
	}

	return &input, true
}

func parseTargetID(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	targetIDParam := vars["targetId"]
	return strconv.ParseInt(targetIDParam, 10, 64)
}
