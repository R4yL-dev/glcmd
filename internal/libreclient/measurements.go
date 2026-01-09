package libreclient

import (
	"context"
	"fmt"
)

// SensorData represents the sensor information from LibreView API.
type SensorData struct {
	SN string `json:"sn"` // Serial number
	A  int    `json:"a"`  // Activation timestamp (Unix)
	PT int    `json:"pt"` // Product type (4 = Libre 3 Plus)
	W  int    `json:"w"`  // Warranty days (not used)
	S  bool   `json:"s"`  // Status (always false, not used)
	LJ bool   `json:"lj"` // Low journey (always false, not used)
}

// ConnectionsResponse represents the response from /llu/connections endpoint.
type ConnectionsResponse struct {
	Data []struct {
		PatientID string `json:"patientId"`
		GlucoseMeasurement struct {
			ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
			Value            float64 `json:"Value"`
			TrendArrow       int     `json:"TrendArrow"`
			TrendMessage     string  `json:"TrendMessage"`
			MeasurementColor int     `json:"MeasurementColor"`
			GlucoseUnits     int     `json:"GlucoseUnits"`
			Timestamp        string  `json:"Timestamp"`
			IsHigh           bool    `json:"isHigh"`
			IsLow            bool    `json:"isLow"`
		} `json:"glucoseMeasurement"`
		Sensor SensorData `json:"sensor"`
	} `json:"data"`
}

// GraphResponse represents the response from /llu/connections/{patientId}/graph endpoint.
type GraphResponse struct {
	Data struct {
		Connection struct {
			GlucoseMeasurement struct {
				ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
				Value            float64 `json:"Value"`
				TrendArrow       int     `json:"TrendArrow"`
				MeasurementColor int     `json:"MeasurementColor"`
				GlucoseUnits     int     `json:"GlucoseUnits"`
				Timestamp        string  `json:"Timestamp"`
				IsHigh           bool    `json:"isHigh"`
				IsLow            bool    `json:"isLow"`
			} `json:"glucoseMeasurement"`
			Sensor SensorData `json:"sensor"`
		} `json:"connection"`
		GraphData []struct {
			FactoryTimestamp string  `json:"FactoryTimestamp"`
			Timestamp        string  `json:"Timestamp"`
			ValueInMgPerDl   int     `json:"ValueInMgPerDl"`
			Value            float64 `json:"Value"`
			MeasurementColor int     `json:"MeasurementColor"`
			GlucoseUnits     int     `json:"GlucoseUnits"`
			IsHigh           bool    `json:"isHigh"`
			IsLow            bool    `json:"isLow"`
			Type             int     `json:"type"`
		} `json:"graphData"`
	} `json:"data"`
}

// GetConnections retrieves the current glucose measurement and patient information.
//
// This endpoint is used for periodic updates (every 5 minutes).
func (c *Client) GetConnections(ctx context.Context, token, accountID string) (*ConnectionsResponse, error) {
	var result ConnectionsResponse
	if err := c.doRequest(ctx, "GET", "/llu/connections", nil, &result, token, accountID); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGraph retrieves historical glucose data (approximately 12 hours).
//
// This endpoint is used for initial data population.
func (c *Client) GetGraph(ctx context.Context, token, accountID, patientID string) (*GraphResponse, error) {
	path := fmt.Sprintf("/llu/connections/%s/graph", patientID)
	var result GraphResponse
	if err := c.doRequest(ctx, "GET", path, nil, &result, token, accountID); err != nil {
		return nil, err
	}
	return &result, nil
}
