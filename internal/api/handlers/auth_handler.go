package handlers

import (
	"encoding/json"
	"flexphish/internal/auth"
	"net/http"
)

type AuthHandler struct {
	service    auth.Service
	jwtService *auth.JWTService
}

func NewAuthHandler(service auth.Service, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		service:    service,
		jwtService: jwtService,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	user, err := h.service.Authenticate(req.Email, req.Password)
	if err != nil {
		JSONResponse(w, http.StatusUnauthorized, map[string]string{
			"error": "invalid credentials",
		})
		return
	}

	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		JSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "could not generate token",
		})
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{
		"token": token,
	})
}
