package headers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/R4yL-dev/glcmd/internal/config"
)

func NewHeaders() *Headers {
	var h Headers

	h.initDefaultHeader()

	return &h
}

func (h *Headers) DefaultHeader() http.Header {
	return h.defaultHeader
}

func (h *Headers) AuthHeader() http.Header {
	return h.authHeader
}

func (h *Headers) BuildAuthHeader(token string, userID string) {
	h.authHeader = h.defaultHeader.Clone()

	hasher := sha256.New()
	hasher.Write([]byte(userID))
	hasherByte := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hasherByte)

	h.authHeader.Set("Authorization", "Bearer "+token)
	h.authHeader.Set("account-id", hashHex)
}

func (h *Headers) initDefaultHeader() {
	h.defaultHeader = config.DefaultHeader.Clone()
	h.authHeader = make(http.Header)
}
