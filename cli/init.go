// Steven Phillips / elimisteve
// 2016.06.05

package cli

import (
	"fmt"
	"strings"

	"github.com/cryptag/cryptag/backend"
)

func InitWebserver(backendName, baseURL, authToken string) error {
	_, err := backend.CreateWebserver(nil, backendName, baseURL, authToken)
	if err != nil && err != backend.ErrConfigExists {
		return err
	}
	return nil
}

func InitSandstorm(backendName, webkey string) error {
	info := strings.SplitN(webkey, "#", 2)
	if len(info) < 2 {
		return fmt.Errorf("Error parsing invalid Sandstorm web key `%s`\n", webkey)
	}
	baseURL, authToken := info[0], info[1]

	return InitWebserver(backendName, baseURL, authToken)
}
