package middleware

import (
	"context"
	"net/http"
	"strconv"

	"flexphish/internal/domain/smtp"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func SMTPOwnershipMiddleware(repo smtp.Repository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value("userID").(int64)

			vars := mux.Vars(r)
			idParam, exists := vars["id"]
			if !exists || idParam == "" {
				next.ServeHTTP(w, r)
				return
			}

			profileID, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				http.Error(w, "invalid smtp profile id", http.StatusBadRequest)
				return
			}

			profile, err := repo.GetByID(profileID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					http.Error(w, "smtp profile not found", http.StatusNotFound)
					return
				}
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			if !profile.IsGlobal {
				if profile.UserId == nil || *profile.UserId != userID {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}

			ctx := context.WithValue(r.Context(), "smtpProfile", profile)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
