package config

import "net/http"

var DefaultHeader = http.Header{
	"User-Agent":   []string{"Mozilla/5.0 (iPhone; CPU OS 17_4.1 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/17.4.1 Mobile/10A5355d Safari/8536.25"},
	"Content-Type": []string{"application/json;charset=UTF-8"},
	"version":      []string{"4.12.0"},
	"product":      []string{"llu.ios"},
}

var LoginURL = "https://api.libreview.io/llu/auth/login"
var ConnectionsURL = "https://api.libreview.io/llu/connections"
var GraphURL = "https://api.libreview.io/llu/connections/%s/graph"
