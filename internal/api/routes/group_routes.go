package routes

import (
	"flexphish/internal/api/handlers"
	"flexphish/internal/api/middleware"
	"flexphish/internal/domain/group"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterGroupRoutes(r *mux.Router, repo group.Repository) {
	handler := handlers.NewGroupHandler(repo)

	groupRouter := r.PathPrefix("/groups").Subrouter()

	groupRouter.HandleFunc("", handler.List).Methods(http.MethodGet, http.MethodOptions)
	groupRouter.HandleFunc("", handler.Create).Methods(http.MethodPost, http.MethodOptions)

	groupRouter.Use(middleware.GroupOwnershipMiddleware(repo))

	groupRouter.HandleFunc("/{id}", handler.Get).Methods(http.MethodGet, http.MethodOptions)
	groupRouter.HandleFunc("/{id}", handler.Update).Methods(http.MethodPut, http.MethodOptions)
	groupRouter.HandleFunc("/{id}", handler.Delete).Methods(http.MethodDelete, http.MethodOptions)

	groupRouter.HandleFunc("/{id}/targets", handler.ListTargets).Methods(http.MethodGet, http.MethodOptions)
	groupRouter.HandleFunc("/{id}/targets", handler.CreateTarget).Methods(http.MethodPost, http.MethodOptions)
	groupRouter.HandleFunc("/{id}/targets/{targetId}", handler.GetTarget).Methods(http.MethodGet, http.MethodOptions)
	groupRouter.HandleFunc("/{id}/targets/{targetId}", handler.UpdateTarget).Methods(http.MethodPut, http.MethodOptions)
	groupRouter.HandleFunc("/{id}/targets/{targetId}", handler.DeleteTarget).Methods(http.MethodDelete, http.MethodOptions)
}
