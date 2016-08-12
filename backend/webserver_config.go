// Steven Phillips / elimisteve
// 2016.02.25

package backend

import "fmt"

type WebserverConfig struct {
	AuthToken string
	BaseURL   string
}

func (wc *WebserverConfig) Valid() error {
	if wc.AuthToken == "" {
		return fmt.Errorf("AuthToken can't be empty")
	}
	if wc.BaseURL == "" {
		return fmt.Errorf("BaseURL can't be empty")
	}
	return nil
}

// Conversions

func WebserverConfigFromMap(m map[string]interface{}) (WebserverConfig, error) {
	var cfg WebserverConfig

	AuthToken, ok := m["AuthToken"].(string)
	if !ok {
		return cfg, fmt.Errorf("Invalid AuthToken '%v'", m["AuthToken"])
	}
	cfg.AuthToken = AuthToken

	BaseURL, ok := m["BaseURL"].(string)
	if !ok {
		return cfg, fmt.Errorf("Invalid BaseURL '%v'", m["BaseURL"])
	}
	cfg.BaseURL = BaseURL

	return cfg, nil
}

func WebserverConfigToMap(cfg WebserverConfig) map[string]interface{} {
	return map[string]interface{}{
		"AuthToken": cfg.AuthToken,
		"BaseURL":   cfg.BaseURL,
	}
}
