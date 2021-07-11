package middleware

import "net/http"

type CorsMiddleware struct {
	allowOrigin string
}

func NewCorsMiddleware(allowOrigin string) *CorsMiddleware {
	return &CorsMiddleware{allowOrigin: allowOrigin}
}

func (m *CorsMiddleware) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", m.allowOrigin)
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "access-control-allow-headers,access-control-allow-methods,access-control-allow-origin,authorization")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}
