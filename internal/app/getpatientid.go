package app

import (
	"encoding/json"
	"fmt"

	"github.com/R4yL-dev/glcmd/internal/config"
	"github.com/R4yL-dev/glcmd/internal/httpreq"
)

func (a *app) getPatientID() error {
	req, err := httpreq.NewHttpReq("GET", config.ConnectionsURL, nil, a.Headers().AuthHeader(), a.ClientHTTP())
	if err != nil {
		return err
	}

	res, err := req.Do()
	if err != nil {
		return err
	}

	var tmp struct {
		Status int `json:"status"`
		Data   []struct {
			PatientID string `json:"patientId"`
		} `json:"data"`
	}

	if err := json.Unmarshal(res, &tmp); err != nil {
		return err
	}

	if tmp.Status != 0 {
		return fmt.Errorf("cannot get patientID: API returned status %d", tmp.Status)
	}

	if len(tmp.Data) == 0 {
		return fmt.Errorf("cannot get patientID: API returned empty data array")
	}

	if err := a.auth.SetPatientID(tmp.Data[0].PatientID); err != nil {
		return err
	}

	return nil
}
