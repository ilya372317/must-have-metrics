package v1

import "net/http"

// StaticHandler file server handler.
func StaticHandler() http.Handler {
	return http.FileServer(http.Dir("static"))
}
