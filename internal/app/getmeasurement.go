package app

import (
	"github.com/R4yL-dev/glcmd/internal/config"
	"github.com/R4yL-dev/glcmd/internal/glucosemeasurement"
	"github.com/R4yL-dev/glcmd/internal/httpreq"
)

func (a *app) GetMeasurement() (*glucosemeasurement.GlucoseMeasurement, error) {
	req, err := httpreq.NewHttpReq("GET", config.ConnectionsURL, nil, a.Headers().AuthHeader(), a.ClientHTTP())
	if err != nil {
		return nil, err
	}
	res, err := req.Do()
	if err != nil {
		return nil, err
	}

	gm, err := glucosemeasurement.NewGlucoseMeasurement(res)
	if err != nil {
		return nil, err
	}

	return gm, nil
}
