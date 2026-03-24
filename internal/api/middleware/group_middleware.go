package middleware

import (
	"context"
	"net/http"
	"strconv"

	"flexphish/internal/domain/group"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func GroupOwnershipMiddleware(repo group.Repository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userID := r.Context().Value("userID").(int64)

			vars := mux.Vars(r)
			idParam, exists := vars["id"]
			if !exists || idParam == "" {
				next.ServeHTTP(w, r)
				return
			}

			groupID, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				http.Error(w, "invalid group id", http.StatusBadRequest)
				return
			}

			g, err := repo.GetByID(groupID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					http.Error(w, "group not found", http.StatusNotFound)
					return
				}

				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			if !g.IsGlobal {
				if g.UserId == nil || *g.UserId != userID {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}

			ctx := context.WithValue(r.Context(), "group", g)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
