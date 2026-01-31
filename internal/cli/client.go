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
	SerialNumber      string   `json:"serialNumber"`
	Activation        string   `json:"activation"`
	ExpiresAt         string   `json:"expiresAt"`
	EndedAt           *string  `json:"endedAt,omitempty"`
	LastMeasurementAt *string  `json:"lastMeasurementAt,omitempty"`
	SensorType        int      `json:"sensorType"`
	DurationDays      int      `json:"durationDays"`
	DaysRemaining     *float64 `json:"daysRemaining,omitempty"`
	DaysElapsed       float64  `json:"daysElapsed"`
	ActualDays        *float64 `json:"actualDays,omitempty"`
	DaysPastExpiry    *float64 `json:"daysPastExpiry,omitempty"`
	IsActive          bool     `json:"isActive"`
	Status            string   `json:"status"`
	IsExpired         bool     `json:"isExpired"`
	IsUnresponsive    bool     `json:"isUnresponsive"`
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

// GetMeasurements fetches glucose measurements with optional filtering
func (c *Client) GetMeasurements(ctx context.Context, params MeasurementParams) (*MeasurementListResponse, error) {
	// Build query string
	path := "/v1/measurements?"
	queryParts := []string{}

	if params.Start != nil {
		queryParts = append(queryParts, fmt.Sprintf("start=%s", params.Start.UTC().Format(time.RFC3339)))
	}
	if params.End != nil {
		queryParts = append(queryParts, fmt.Sprintf("end=%s", params.End.UTC().Format(time.RFC3339)))
	}
	if params.Limit > 0 {
		queryParts = append(queryParts, fmt.Sprintf("limit=%d", params.Limit))
	}

	for i, part := range queryParts {
		if i > 0 {
			path += "&"
		}
		path += part
	}

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to glcore at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result MeasurementListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetStatistics fetches glucose statistics for a time period
func (c *Client) GetStatistics(ctx context.Context, start, end *time.Time) (*StatisticsResponse, error) {
	// Build query string
	path := "/v1/measurements/stats"
	queryParts := []string{}

	if start != nil {
		queryParts = append(queryParts, fmt.Sprintf("start=%s", start.UTC().Format(time.RFC3339)))
	}
	if end != nil {
		queryParts = append(queryParts, fmt.Sprintf("end=%s", end.UTC().Format(time.RFC3339)))
	}

	if len(queryParts) > 0 {
		path += "?"
		for i, part := range queryParts {
			if i > 0 {
				path += "&"
			}
			path += part
		}
	}

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to glcore at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result StatisticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetSensorHistory fetches sensor history with optional filtering
func (c *Client) GetSensorHistory(ctx context.Context, params SensorHistoryParams) (*SensorListResponse, error) {
	path := "/v1/sensors/history?"
	queryParts := []string{}

	if params.Start != nil {
		queryParts = append(queryParts, fmt.Sprintf("start=%s", params.Start.UTC().Format(time.RFC3339)))
	}
	if params.End != nil {
		queryParts = append(queryParts, fmt.Sprintf("end=%s", params.End.UTC().Format(time.RFC3339)))
	}
	if params.Limit > 0 {
		queryParts = append(queryParts, fmt.Sprintf("limit=%d", params.Limit))
	}

	for i, part := range queryParts {
		if i > 0 {
			path += "&"
		}
		path += part
	}

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to glcore at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result SensorListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetSensorStatistics fetches sensor lifecycle statistics
func (c *Client) GetSensorStatistics(ctx context.Context) (*SensorStatisticsResponse, error) {
	resp, err := c.get(ctx, "/v1/sensors/stats")
	if err != nil {
		return nil, fmt.Errorf("cannot connect to glcore at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result SensorStatisticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	return c.httpClient.Do(req)
}
