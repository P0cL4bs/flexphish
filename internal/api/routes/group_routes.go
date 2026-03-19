package routes

import (
	"flexphish/internal/api/handlers"
	"flexphish/internal/api/middleware"
	"flexphish/internal/domain/group"

	"github.com/gorilla/mux"
)

func RegisterGroupRoutes(r *mux.Router, repo group.Repository) {
	handler := handlers.NewGroupHandler(repo)

	groupRouter := r.PathPrefix("/groups").Subrouter()

	groupRouter.HandleFunc("", handler.List).Methods("GET")
	groupRouter.HandleFunc("", handler.Create).Methods("POST")

	groupRouter.Use(middleware.GroupOwnershipMiddleware(repo))

	groupRouter.HandleFunc("/{id}", handler.Get).Methods("GET")
	groupRouter.HandleFunc("/{id}", handler.Update).Methods("PUT")
	groupRouter.HandleFunc("/{id}", handler.Delete).Methods("DELETE")

	groupRouter.HandleFunc("/{id}/targets", handler.ListTargets).Methods("GET")
	groupRouter.HandleFunc("/{id}/targets", handler.CreateTarget).Methods("POST")
	groupRouter.HandleFunc("/{id}/targets/{targetId}", handler.GetTarget).Methods("GET")
	groupRouter.HandleFunc("/{id}/targets/{targetId}", handler.UpdateTarget).Methods("PUT")
	groupRouter.HandleFunc("/{id}/targets/{targetId}", handler.DeleteTarget).Methods("DELETE")
}
