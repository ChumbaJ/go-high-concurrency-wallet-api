// Package middleware
package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"test-psql/pkg/logger"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error(fmt.Sprintf("panic recovered: %v method=%s path=%s", rec, r.Method, r.URL.Path))
				logger.Error(fmt.Sprintf("stacktrace:\n%s", debug.Stack()))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
