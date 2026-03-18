package handlers

import (
	"encoding/json"
	"flexphish/internal/config"
	"net/http"

	"github.com/knadh/koanf"
)

type ConfigHandler struct {
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

func (h *ConfigHandler) List(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config

	if err := config.Get().UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		http.Error(w, "failed to unmarshal config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(cfg); err != nil {
		http.Error(w, "failed to encode config to JSON", http.StatusInternalServerError)
		return
	}
}

type updateConfigRequest map[string]interface{}

func (h *ConfigHandler) Update(w http.ResponseWriter, r *http.Request) {

	var req updateConfigRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json body",
		})
		return
	}

	for key, value := range req {

		if err := config.SetConfigField(key, value); err != nil {

			JSONResponse(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}
	}

	JSONResponse(w, http.StatusOK, map[string]string{
		"status": "config updated",
	})
}
