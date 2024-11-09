package middleware

import "net/http"

type Step func(http.HandlerFunc) http.HandlerFunc

// Prepend the middlewares to the handler in the order they are provided,
// and return the resulting (chained) handler.
func Chain(h http.HandlerFunc, m ...Step) http.HandlerFunc {
	if len(m) == 0 {
		return h
	}
	handler := h
	for i := len(m) - 1; i >= 0; i-- {
		handler = m[i](handler)
	}
	return handler
}
