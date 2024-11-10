package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

func CORSAllowAllOrigin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the header to allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next(w, r)
	}
}

type cors struct {
	origin  string
	headers string
	methods string
	maxAge  string
}

type corsOption func(*cors)

func AllowOrigin(origin string) corsOption {
	return func(c *cors) {
		c.origin = origin
	}
}

func AllowMethods(methods ...string) corsOption {
	return func(c *cors) {
		c.methods = strings.Join(methods, ",")
	}
}

func AllowHeaders(headers ...string) corsOption {
	return func(c *cors) {
		c.headers = strings.Join(headers, ",")
	}
}

func MaxAge(d time.Duration) corsOption {
	return func(c *cors) {
		c.maxAge = strconv.Itoa(int(d.Seconds()))
	}
}

// Default allows all methods, all headers, and sets the max age to 1 day.
func CORS(opts ...corsOption) Step {
	cors := &cors{
		origin:  "*",
		methods: "*",
		headers: "*",
		maxAge:  "86400",
	}
	for _, opt := range opts {
		opt(cors)
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", cors.origin)
			if r.Method == http.MethodOptions {
				// Set the header to allow all origins
				w.Header().Set("Access-Control-Allow-Methods", cors.methods)
				w.Header().Set("Access-Control-Allow-Headers", cors.headers)
				w.Header().Set("Access-Control-Max-Age", cors.maxAge)
			}
			next(w, r)
		}
	}
}
