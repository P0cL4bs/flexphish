package routes

import (
	"flexphish/internal/api/handlers"
	"flexphish/internal/api/middleware"
	"flexphish/internal/domain/smtp"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterSMTPRoutes(r *mux.Router, repo smtp.Repository) {
	handler := handlers.NewSMTPHandler(repo)

	smtpRouter := r.PathPrefix("/smtp-profiles").Subrouter()

	smtpRouter.HandleFunc("", handler.List).Methods(http.MethodGet, http.MethodOptions)
	smtpRouter.HandleFunc("", handler.Create).Methods(http.MethodPost, http.MethodOptions)
	smtpRouter.HandleFunc("/test", handler.TestConnection).Methods(http.MethodPost, http.MethodOptions)

	smtpRouter.Use(middleware.SMTPOwnershipMiddleware(repo))

	smtpRouter.HandleFunc("/{id}", handler.Get).Methods(http.MethodGet, http.MethodOptions)
	smtpRouter.HandleFunc("/{id}", handler.Update).Methods(http.MethodPut, http.MethodOptions)
	smtpRouter.HandleFunc("/{id}", handler.Delete).Methods(http.MethodDelete, http.MethodOptions)
}
