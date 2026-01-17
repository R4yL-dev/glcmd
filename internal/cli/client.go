package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client wraps HTTP calls to the glcore API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new CLI client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GlucoseReading represents the glucose data returned by the API
type GlucoseReading struct {
	Value          float64   `json:"value"`
	ValueInMgPerDl int       `json:"valueInMgPerDl"`
	TrendArrow     *int      `json:"trendArrow,omitempty"`
	MeasurementColor int     `json:"measurementColor"`
	IsHigh         bool      `json:"isHigh"`
	IsLow          bool      `json:"isLow"`
	Timestamp      time.Time `json:"timestamp"`
	GlucoseUnits   int       `json:"glucoseUnits"`
}

// SensorInfo represents the sensor data returned by the API
type SensorInfo struct {
	SerialNumber  string   `json:"serialNumber"`
	Activation    string   `json:"activation"`
	ExpiresAt     string   `json:"expiresAt"`
	DaysRemaining *float64 `json:"daysRemaining,omitempty"`
	DaysElapsed   float64  `json:"daysElapsed"`
	IsActive      bool     `json:"isActive"`
}

// GetLatestGlucose fetches the latest glucose reading
func (c *Client) GetLatestGlucose(ctx context.Context) (*GlucoseReading, error) {
	resp, err := c.get(ctx, "/v1/measurements/latest")
	if err != nil {
		return nil, fmt.Errorf("cannot connect to glcore at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no glucose readings available")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Data *GlucoseReading `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Data == nil {
		return nil, fmt.Errorf("no glucose readings available")
	}

	return result.Data, nil
}

// GetCurrentSensor fetches the current sensor info
func (c *Client) GetCurrentSensor(ctx context.Context) (*SensorInfo, error) {
	resp, err := c.get(ctx, "/v1/sensors")
	if err != nil {
		return nil, fmt.Errorf("cannot connect to glcore at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Current *SensorInfo `json:"current"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Data.Current == nil {
		return nil, fmt.Errorf("no active sensor found")
	}

	return result.Data.Current, nil
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	return c.httpClient.Do(req)
}
