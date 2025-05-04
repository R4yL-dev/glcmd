package httpreq

import (
	"net/http"
	"net/url"
)

type httpReq struct {
	method  string
	url     *url.URL
	payload []byte
	headers http.Header
	client  *http.Client
}
