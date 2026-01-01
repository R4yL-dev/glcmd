package libreclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// decodeJSONResponse decodes a JSON response body into the provided result.
func decodeJSONResponse(resp *http.Response, result interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return fmt.Errorf("empty response body")
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}
