package middleware

import (
	"context"
	"net/http"
	"strconv"

	"flexphish/internal/domain/campaign"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func CampaignOwnershipMiddleware(repo campaign.Repository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userID := r.Context().Value("userID").(int64)

			vars := mux.Vars(r)
			idParam, exists := vars["id"]
			if !exists || idParam == "" {
				next.ServeHTTP(w, r)
				return
			}

			campaignID, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				http.Error(w, "invalid campaign id", http.StatusBadRequest)
				return
			}

			camp, err := repo.GetByID(campaignID, userID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					http.Error(w, "campaign not found", http.StatusNotFound)
					return
				}

				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), "campaign", camp)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func PreventModificationIfActive() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			value := r.Context().Value("campaign")
			if value == nil {
				next.ServeHTTP(w, r)
				return
			}

			camp := value.(*campaign.Campaign)

			if camp.Status == campaign.StatusActive ||
				camp.Status == campaign.StatusCompleted {

				http.Error(w,
					"cannot modify active or completed campaign",
					http.StatusBadRequest,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
