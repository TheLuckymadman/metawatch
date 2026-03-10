// Middlewares.
package middleware

import (
	"net/http"
)

type (
	Middleware   func(http.HandlerFunc) http.HandlerFunc
	responseData struct {
		status int
		size   int
	}
)

func MiddlewareConveyor(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}
