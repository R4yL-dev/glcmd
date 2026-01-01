package libreclient

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// AuthResponse represents the authentication response from LibreView API.
type AuthResponse struct {
	Data struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
		AuthTicket struct {
			Token    string `json:"token"`
			Expires  int64  `json:"expires"`
			Duration int64  `json:"duration"`
		} `json:"authTicket"`
	} `json:"data"`
}

// AuthCredentials holds the credentials for authentication.
type AuthCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Authenticate authenticates with the LibreView API and returns the auth token and user ID.
func (c *Client) Authenticate(ctx context.Context, email, password string) (token, userID, accountID string, err error) {
	creds := AuthCredentials{
		Email:    email,
		Password: password,
	}

	var resp AuthResponse
	// No auth needed for login endpoint (empty strings for token/accountID)
	if err := c.doRequest(ctx, "POST", "/llu/auth/login", creds, &resp, "", ""); err != nil {
		return "", "", "", err
	}

	// Calculate account ID (SHA256 hash of user ID)
	hasher := sha256.New()
	hasher.Write([]byte(resp.Data.User.ID))
	hashBytes := hasher.Sum(nil)
	accountID = hex.EncodeToString(hashBytes)

	return resp.Data.AuthTicket.Token, resp.Data.User.ID, accountID, nil
}
