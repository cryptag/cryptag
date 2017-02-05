// Steve Phillips / elimisteve
// 2017.01.16

package share

import (
	"log"
	"net/http"
	"strings"

	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/tor"
	"github.com/cryptag/cryptag/types"
)

type Client struct {
	ServerBaseURL string
	AuthToken     string
	Client        *http.Client
	UsingTor      bool
}

func NewClient(serverBaseURL string) *Client {
	serverBaseURL = cleanBase(serverBaseURL)

	useTor := cryptag.UseTor || strings.HasSuffix(serverBaseURL, ".onion")

	return &Client{
		ServerBaseURL: serverBaseURL,
		Client:        getClient(useTor),
		UsingTor:      useTor,
	}
}

func cleanBase(serverBaseURL string) string {
	serverBaseURL = strings.TrimRight(serverBaseURL, "/")

	if serverBaseURL == "" {
		serverBaseURL = DefaultServerURL
	}

	if !strings.HasPrefix(serverBaseURL, "http://") &&
		!strings.HasPrefix(serverBaseURL, "https://") {

		if strings.HasSuffix(serverBaseURL, ".onion") {
			serverBaseURL = "http://" + serverBaseURL
		} else {
			serverBaseURL = "https://" + serverBaseURL
		}
	}

	return serverBaseURL
}

// getClient returns an appropriate HTTP client judging by whether
// useTor is set to true.
func getClient(useTor bool) *http.Client {
	if useTor {
		client, err := tor.NewClient()
		if err != nil {
			log.Printf("Error creating Tor-capable client %v\n", err)
			return http.DefaultClient
		}
		if types.Debug {
			log.Println("Returning Tor-capable HTTP client...")
		}
		return client
	}

	// TODO: Use client that eventually times out
	return http.DefaultClient
}
