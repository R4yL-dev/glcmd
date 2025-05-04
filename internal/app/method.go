package app

import (
	"net/http"

	"github.com/R4yL-dev/glcmd/internal/auth"
	"github.com/R4yL-dev/glcmd/internal/credentials"
	"github.com/R4yL-dev/glcmd/internal/headers"
)

type app struct {
	credentials *credentials.Credentials
	auth        *auth.Auth
	headers     *headers.Headers
	clientHTTP  *http.Client
}
