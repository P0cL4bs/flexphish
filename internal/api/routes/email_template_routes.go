package routes

import (
	"flexphish/internal/api/handlers"
	"flexphish/internal/api/middleware"
	"flexphish/internal/domain/template"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterEmailTemplateRoutes(r *mux.Router, repo template.EmailTemplateRepository) {
	handler := handlers.NewEmailTemplateHandler(repo)

	emailTemplateRouter := r.PathPrefix("/email-templates").Subrouter()

	emailTemplateRouter.HandleFunc("", handler.List).Methods(http.MethodGet, http.MethodOptions)
	emailTemplateRouter.HandleFunc("", handler.Create).Methods(http.MethodPost, http.MethodOptions)

	emailTemplateRouter.Use(middleware.EmailTemplateOwnershipMiddleware(repo))

	emailTemplateRouter.HandleFunc("/{id}", handler.Get).Methods(http.MethodGet, http.MethodOptions)
	emailTemplateRouter.HandleFunc("/{id}", handler.Update).Methods(http.MethodPut, http.MethodOptions)
	emailTemplateRouter.HandleFunc("/{id}", handler.Delete).Methods(http.MethodDelete, http.MethodOptions)
}
