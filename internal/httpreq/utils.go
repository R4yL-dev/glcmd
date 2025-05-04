package httpreq

import (
	"net"
	"net/url"
	"strings"
)

func isValidMethod(method string) bool {
	return validMethods[method]
}

func isValidURL(u *url.URL) bool {
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	if u.Host == "" {
		return false
	}

	if !hasTLD(u) {
		return false
	}
	if err := urlLookup(u); err != nil {
		return false
	}
	return true
}

func hasTLD(u *url.URL) bool {
	host := u.Hostname()
	return strings.Contains(host, ".")
}

func urlLookup(u *url.URL) error {
	if _, err := net.LookupHost(u.Hostname()); err != nil {
		return err
	}
	return nil
}
