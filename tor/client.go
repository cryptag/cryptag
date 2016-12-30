// Steven Phillips / elimisteve
// 2016.04.18

package tor

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/net/proxy"
)

var (
	// TorBrowserBundleProxyURL is the URL to the local SOCKS5 proxy
	// started by the Tor Browser Bundle. Set 'ProxyURL' to this if
	// you're using -- you guessed it -- the Tor Browser Bundle.
	TorBrowserBundleProxyURL = "socks5://127.0.0.1:9150"

	// TorServiceProxyURL is the URL to the local SOCKS5 proxy started
	// by the Tor service/daemon. Set 'ProxyURL' to this if you have
	// the Tor service installed on your machine.
	TorServiceProxyURL = "socks5://127.0.0.1:9050"

	// ProxyURL is the URL of the local SOCKS5 proxy used to make
	// requests over Tor. Defaults to TorBrowserBundleProxyURL.
	//
	// Local proxy port used to make requests over Tor can be set with
	// TOR_PORT environment variable (in which case ProxyURL gets set
	// to "127.0.0.1:${TOR_PORT}").
	ProxyURL = TorBrowserBundleProxyURL
)

func init() {
	if p := os.Getenv("TOR_PORT"); p != "" {
		ProxyURL = "socks5://127.0.0.1:" + p
	}
}

// NewClient returns an HTTP client that does requests through Tor.
func NewClient() (*http.Client, error) {
	proxyURL, err := url.Parse(ProxyURL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing proxy URL: %v", err)
	}

	// Thank you https://gist.github.com/Yawning/bac58e08a05fc378a8cc
	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{Dial: dialer.Dial}

	client := &http.Client{Transport: transport}

	return client, nil
}
