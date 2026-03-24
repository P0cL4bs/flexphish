package routes

import (
	"net/http"

	"flexphish/internal/api/handlers"

	"github.com/gorilla/mux"
)

func RegisterTemplateRoutes(router *mux.Router, handler *handlers.TemplateHandler) {
	api := router.PathPrefix("/templates").Subrouter()

	api.HandleFunc("", handler.List).Methods(http.MethodGet, http.MethodOptions)
	api.HandleFunc("", handler.Create).Methods(http.MethodPost, http.MethodOptions)
	api.HandleFunc("/import", handler.ImportZip).Methods(http.MethodPost, http.MethodOptions)
	api.HandleFunc("", handler.Update).Methods(http.MethodPut, http.MethodOptions)
	api.HandleFunc("", handler.Delete).Methods(http.MethodDelete, http.MethodOptions)
	api.HandleFunc("/{filename}/clone", handler.Clone).Methods(http.MethodPost, http.MethodOptions)

	api.HandleFunc("/{filename}", handler.GetByFilename).Methods(http.MethodGet, http.MethodOptions)
	api.HandleFunc("/{filename}/export", handler.ExportZip).Methods(http.MethodGet, http.MethodOptions)
	api.HandleFunc("/{filename}/metadata", handler.GetMetadataByFilename).Methods(http.MethodGet, http.MethodOptions)
}

func RegisterHtmlfilesRoutes(router *mux.Router, handler *handlers.HtmlFileHandler) {
	api_htmlfiles := router.PathPrefix("/templates/{filename}/html-files").Subrouter()

	api_htmlfiles.HandleFunc("", handler.List).Methods(http.MethodGet, http.MethodOptions)
	api_htmlfiles.HandleFunc("", handler.Create).Methods(http.MethodPost, http.MethodOptions)
	api_htmlfiles.HandleFunc("", handler.Update).Methods(http.MethodPut, http.MethodOptions)
	api_htmlfiles.HandleFunc("", handler.Delete).Methods(http.MethodDelete, http.MethodOptions)
}

func RegisterStaticFilesRoutes(router *mux.Router, handler *handlers.StaticFileHandler) {
	api_staticfiles := router.PathPrefix("/templates/{filename}/static-files").Subrouter()

	api_staticfiles.HandleFunc("", handler.List).Methods(http.MethodGet, http.MethodOptions)
	api_staticfiles.HandleFunc("", handler.Create).Methods(http.MethodPost, http.MethodOptions)
	api_staticfiles.HandleFunc("", handler.Update).Methods(http.MethodPut, http.MethodOptions)
	api_staticfiles.HandleFunc("", handler.Delete).Methods(http.MethodDelete, http.MethodOptions)
}
