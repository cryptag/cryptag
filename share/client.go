// Steve Phillips / elimisteve
// 2017.01.16

package share

import (
	"log"
	"net/http"

	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/tor"
	"github.com/cryptag/cryptag/types"
)

// getClient returns an appropriate HTTP client judging by whether
// cryptag.UseTor is set to true.
func getClient() *http.Client {
	if cryptag.UseTor {
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
