package http

import (
	"net/http"
)

// InitRoutes initializes routes.
func InitRoutes() {
	http.HandleFunc("/chat", handleWS)
}
