package routes

import (
	"flexphish/internal/api/handlers"
	"flexphish/internal/api/middleware"
	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/template"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterCampaignRoutes(
	router *mux.Router,
	repo campaign.Repository,
	trepo template.TemplateRepository,
	jwtMiddleware mux.MiddlewareFunc,
) {

	handler := handlers.NewCampaignHandler(repo, trepo)

	campaignRouter := router.PathPrefix("/campaigns").Subrouter()
	campaignRouter.HandleFunc("/analytics", handler.Analytics).Methods(http.MethodGet, http.MethodOptions)
	campaignRouter.HandleFunc("", handler.List).Methods(http.MethodGet, http.MethodOptions)
	campaignRouter.HandleFunc("", handler.Create).Methods(http.MethodPost, http.MethodOptions)

	idRouter := campaignRouter.PathPrefix("/{id}").Subrouter()
	idRouter.Use(middleware.CampaignOwnershipMiddleware(repo))

	idRouter.HandleFunc("", handler.GetByID).Methods(http.MethodGet, http.MethodOptions)
	idRouter.HandleFunc("", handler.Update).Methods(http.MethodPut, http.MethodOptions)
	idRouter.HandleFunc("", handler.Delete).Methods(http.MethodDelete, http.MethodOptions)

	idRouter.HandleFunc("/start", handler.Start).Methods(http.MethodPost, http.MethodOptions)
	idRouter.HandleFunc("/stop", handler.Stop).Methods(http.MethodPost, http.MethodOptions)
	idRouter.HandleFunc("/archive", handler.Archive).Methods(http.MethodPost, http.MethodOptions)

	idRouter.HandleFunc("/results/{result_id}", handler.DeleteResult).Methods(http.MethodDelete, http.MethodOptions)
}
