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

func CreateSandstormWebserver(key []byte, bkName, webkey string) (*WebserverBackend, error) {
	var goodKey *[32]byte

	if len(key) > 0 {
		var err error
		goodKey, err = cryptag.ConvertKey(key)
		if err != nil {
			return nil, fmt.Errorf("Error converting key: %v", err)
		}
	}

	cfg := &Config{
		Name:   bkName,
		Type:   TypeSandstorm,
		Key:    goodKey,
		Custom: SandstormWebKeyToMap(webkey),
	}

	if err := cfg.Canonicalize(); err != nil {
		return nil, err
	}

	if err := cfg.Save(cryptag.BackendPath); err != nil {
		return nil, err
	}

	return SandstormFromConfig(cfg)
}
