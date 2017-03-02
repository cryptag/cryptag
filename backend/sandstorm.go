// Steve Phillips / elimisteve
// 2017.02.18

package backend

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cryptag/cryptag"
)

var (
	ErrNilCustom = errors.New("Field 'Custom' cannot be nil!")
)

func SandstormFromConfig(cfg *Config) (*WebserverBackend, error) {
	if cfg.Key == nil {
		return nil, cryptag.ErrNilKey
	}
	if cfg.Custom == nil {
		return nil, ErrNilCustom
	}

	webkey := fmt.Sprintf("%v", cfg.Custom["WebKey"])

	info := strings.SplitN(webkey, "#", 2)
	if len(info) < 2 {
		return nil, fmt.Errorf("Error parsing invalid Sandstorm web key `%s`\n", webkey)
	}
	baseURL, authToken := info[0], info[1]

	return NewWebserverBackend((*cfg.Key)[:], cfg.Name, baseURL, authToken)
}
