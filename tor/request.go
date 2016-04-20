// Steven Phillips / elimisteve
// 2016.04.17

package tor

import (
	"fmt"
	"io"
	"net/http"
)

const (
	torBrowserUserAgent = "Mozilla/5.0 (Windows NT 6.1; rv:38.0) Gecko/20100101 Firefox/38.0"
)

// DoRequest performs the given request through Tor, using
// socks5://127.0.0.1:9150 (change ProxyURL to use different
// proxy). If client is nil, a client is created (with NewClient).
func DoRequest(client *http.Client, method, url string, body io.Reader) (*http.Response, error) {
	if client == nil {
		var err error
		client, err = NewClient()
		if err != nil {
			return nil, err
		}
	}

	req, err := NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

// Get does a GET request to the given URL over Tor.
func Get(url string) (*http.Response, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return DoRequest(client, "GET", url, nil)
}

// Post does a POST request to the given URL over Tor.
func Post(url string, body io.Reader) (*http.Response, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return DoRequest(client, "GET", url, body)
}

// NewRequest creates a *http.Request with the User-Agent set to what
// Tor Browser uses.
func NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("Error creating %s request: %v", method, err)
	}

	// Change User Agent string (default: "Go http package")
	req.Header.Set("User-Agent", torBrowserUserAgent)

	return req, nil
}
