package middleware

import (
	"context"
	"net/http"
	"strconv"

	"flexphish/internal/domain/template"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func EmailTemplateOwnershipMiddleware(repo template.EmailTemplateRepository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value("userID").(int64)

			vars := mux.Vars(r)
			idParam, exists := vars["id"]
			if !exists || idParam == "" {
				next.ServeHTTP(w, r)
				return
			}

			templateID, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				http.Error(w, "invalid email template id", http.StatusBadRequest)
				return
			}

			emailTemplate, err := repo.GetByID(templateID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					http.Error(w, "email template not found", http.StatusNotFound)
					return
				}
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			if !emailTemplate.IsGlobal {
				if emailTemplate.UserId == nil || *emailTemplate.UserId != userID {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}

			ctx := context.WithValue(r.Context(), "emailTemplate", emailTemplate)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
