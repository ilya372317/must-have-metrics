package middleware

import (
	"net/http"
)

// Middleware base signature for all middlewares.
type Middleware func(handler http.Handler) http.Handler
