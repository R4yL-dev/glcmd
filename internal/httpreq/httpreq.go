package httpreq

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/R4yL-dev/glcmd/internal/config"
)

var validMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodPost:    true,
	http.MethodPut:     false,
	http.MethodDelete:  false,
	http.MethodPatch:   false,
	http.MethodHead:    false,
	http.MethodOptions: false,
}

type httpReq struct {
	method  string
	url     *url.URL
	payload []byte
	headers http.Header
	client  *http.Client
}

func NewHttpReq(method string, url string, payload []byte, headers http.Header, client *http.Client) (*httpReq, error) {
	var newReq httpReq

	if err := newReq.SetMethod(method); err != nil {
		return nil, err
	}

	if err := newReq.SetUrl(url); err != nil {
		return nil, err
	}

	if payload != nil {
		newReq.SetPayload(payload)
	}

	newReq.SetHeaders(headers)

	newReq.SetClient(client)

	return &newReq, nil
}

func (req *httpReq) Method() string {
	return req.method
}

func (req *httpReq) Url() *url.URL {
	return req.url
}

func (req *httpReq) Payload() []byte {
	return req.payload
}

func (req *httpReq) Headers() http.Header {
	req.initHeaders()
	return req.headers
}

func (req *httpReq) Client() *http.Client {
	return req.client
}

func (req *httpReq) SetMethod(m string) error {
	if m == "" {
		return fmt.Errorf("method cannot be empty")
	}
	if !isValidMethod(m) {
		return fmt.Errorf("method is invalid: %s", m)
	}
	req.method = m
	return nil
}

func (req *httpReq) SetUrl(t string) error {
	if t == "" {
		return fmt.Errorf("url cannot be empty")
	}
	var newUrl *url.URL
	newUrl, err := url.Parse(t)
	if err != nil {
		return fmt.Errorf("target cannot be parsed: %s: %v", t, err)
	}
	if !isValidURL(newUrl) {
		return fmt.Errorf("target in invalid: %s", t)
	}
	req.url = newUrl
	return nil
}

func (req *httpReq) SetPayload(p []byte) {
	req.payload = p
}

func (req *httpReq) SetHeaders(headers http.Header) {
	req.initHeaders()

	for k, v := range headers {
		req.headers[k] = append([]string(nil), v...)
	}
}

func (req *httpReq) SetClient(c *http.Client) {
	if c == nil {
		req.client = &http.Client{}
	} else {
		req.client = c
	}
}

func (req *httpReq) initHeaders() {
	if req.headers == nil {
		req.headers = make(http.Header)
		for k, v := range config.DefaultHeader {
			req.headers[k] = append([]string(nil), v...)
		}
	}
}

func (req *httpReq) IsValid() bool {
	if req.method == "" || !isValidMethod(req.method) {
		return false
	}
	if req.url == nil || !isValidURL(req.url) {
		return false
	}
	if req.headers == nil {
		return false
	}
	if req.client == nil {
		return false
	}
	return true
}

func (req *httpReq) ToRequest() (*http.Request, error) {
	if !req.IsValid() {
		return nil, fmt.Errorf("HttpReq invalid")
	}
	var httpRequest *http.Request
	var err error
	if req.payload != nil {
		httpRequest, err = http.NewRequest(req.method, req.url.String(), bytes.NewBuffer(req.payload))
	} else {
		httpRequest, err = http.NewRequest(req.method, req.url.String(), nil)
	}
	if err != nil {
		return nil, err
	}

	for k, v := range req.Headers() {
		for _, val := range v {
			httpRequest.Header.Add(k, val)
		}
	}
	return httpRequest, nil
}

func (req *httpReq) Do() ([]byte, error) {
	httpReq, err := req.ToRequest()
	if err != nil {
		return nil, err
	}

	res, err := req.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error while processing http request: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected HTTP status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error during parsing body: %s", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	return body, nil
}
