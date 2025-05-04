package app

import (
	"github.com/R4yL-dev/glcmd/internal/auth"
	"github.com/R4yL-dev/glcmd/internal/config"
	"github.com/R4yL-dev/glcmd/internal/httpreq"
)

func (a *app) ensureAuth() error {
	if !a.auth.IsAuth() {
		payload, err := a.credentials.ToJSON()
		if err != nil {
			return err
		}

		req, err := httpreq.NewHttpReq("POST", config.LoginURL, payload, a.headers.DefaultHeader(), a.ClientHTTP())
		if err != nil {
			return err
		}

		res, err := req.Do()
		if err != nil {
			return err
		}

		newAuth, err := auth.NewAuth(res)
		if err != nil {
			return err
		}

		a.auth = newAuth

		a.headers.BuildAuthHeader(a.auth.Ticket().Token(), a.auth.UserID())

		if err := a.getPatientID(); err != nil {
			return err
		}
	}
	return nil
}
