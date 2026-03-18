package routes

import (
	"flexphish/internal/api/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterConfigRoutes(router *mux.Router, handler *handlers.ConfigHandler) {

	api := router.PathPrefix("/configs").Subrouter()

	api.HandleFunc("", handler.List).Methods(http.MethodGet, http.MethodOptions)
	api.HandleFunc("", handler.Update).Methods(http.MethodPut, http.MethodOptions)
}
