package headers

import "net/http"

type Headers struct {
	defaultHeader http.Header
	authHeader    http.Header
}
