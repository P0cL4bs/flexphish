package routes

import "net/http"

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("X-Frame-Options", "DENY")           // Block iframe embedding
		w.Header().Set("X-Content-Type-Options", "nosniff") // Prevent MIME sniffing
		w.Header().Set("X-XSS-Protection", "1; mode=block") // XSS filter
		w.Header().Set("Referrer-Policy", "same-origin")    // Limit referrer leakage

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers",
			"Authorization, Content-Type, Accept, X-Requested-With, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
