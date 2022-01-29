package middleware

import (
	"github.com/golauth/golauth/domain/usecase/token"
	"net/http"
)

type SecurityMiddleware struct {
	service       token.UseCase
	validateToken token.ValidateToken
	publicURI     map[string]bool
}

func NewSecurityMiddleware(service token.UseCase, validateToken token.ValidateToken, pathPrefix string) *SecurityMiddleware {
	return &SecurityMiddleware{
		service:       service,
		validateToken: validateToken,
		publicURI: map[string]bool{
			pathPrefix + "/token":       true,
			pathPrefix + "/check_token": true,
			pathPrefix + "/signup":      true,
		},
	}
}

func (s *SecurityMiddleware) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURI := r.RequestURI
		if s.isPrivateURI(requestURI) {
			t, err := s.service.ExtractToken(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = s.validateToken.Execute(t)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s SecurityMiddleware) isPrivateURI(requestURI string) bool {
	_, contains := s.publicURI[requestURI]
	return !contains
}
