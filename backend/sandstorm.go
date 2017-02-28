// Steve Phillips / elimisteve
// 2017.02.18

package backend

import (
	"errors"
	"fmt"
)

var (
	ErrNilCustom = errors.New("Field 'Custom' cannot be nil!")
)

func SandstormFromConfig(cfg *Config) (*WebserverBackend, error) {
	if cfg.Custom == nil {
		return nil, ErrNilCustom
	}

	args := []string{str(cfg.Custom["WebKey"])}
	if cfg.Custom["WebKey"] == "" {
		args = []string{str(cfg.Custom["BaseURL"]), str(cfg.Custom["AuthToken"])}
	}

	bk, err := Create(TypeWebserver, cfg.Name, args)
	if err != nil {
		return nil, err
	}
	wbk, ok := bk.(*WebserverBackend)
	if !ok {
		return nil, fmt.Errorf("Error turning Backend into *WebserverBackend! It's a %T", bk)
	}
	return wbk, nil
}

func str(obj interface{}) string {
	if obj == nil {
		return ""
	}
	return fmt.Sprintf("%s", obj)
}
