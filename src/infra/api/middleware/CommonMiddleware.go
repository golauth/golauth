package middleware

import "net/http"

type CommonMiddleware struct {
}

func NewCommonMiddleware() *CommonMiddleware {
	return &CommonMiddleware{}
}

func (m *CommonMiddleware) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
