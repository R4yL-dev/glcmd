package app

import (
	"fmt"
	"net/http"

	"github.com/R4yL-dev/glcmd/internal/auth"
	"github.com/R4yL-dev/glcmd/internal/credentials"
	"github.com/R4yL-dev/glcmd/internal/headers"
)

func NewApp(email string, password string, client *http.Client) (*app, error) {
	creds, err := credentials.NewCredentials(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %v", err)
	}

	h := headers.NewHeaders()

	if client == nil {
		client = &http.Client{}
	}

	newApp := &app{
		credentials: creds,
		auth:        &auth.Auth{},
		headers:     h,
		clientHTTP:  client,
	}

	if err := newApp.ensureAuth(); err != nil {
		return nil, err
	}

	return newApp, nil
}

func (a *app) Credentials() *credentials.Credentials {
	return a.credentials
}

func (a *app) Auth() *auth.Auth {
	return a.auth
}

func (a *app) Headers() *headers.Headers {
	return a.headers
}

func (a *app) ClientHTTP() *http.Client {
	return a.clientHTTP
}
