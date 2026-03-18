package routes

import (
	"net/http"

	"flexphish/internal/application/controller"
)

func RegisterRuntimeRoutes(
	mux *http.ServeMux,
	runtimeController *controller.CampaignRuntimeController,
) {
	runtimeController.RegisterRoutes(mux)
}
