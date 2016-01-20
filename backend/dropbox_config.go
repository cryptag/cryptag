// Steven Phillips / elimisteve
// 2016.01.20

package backend

import (
	"fmt"
)

type DropboxConfig struct {
	AppKey      string
	AppSecret   string
	AccessToken string
	BasePath    string // e.g., "/cryptag_folder_in_dropbox_root"
}

func (dc *DropboxConfig) Valid() error {
	if dc.AppKey == "" {
		return fmt.Errorf("Invalid AppKey '%v'", dc.AppKey)
	}
	if dc.AppSecret == "" {
		return fmt.Errorf("Invalid AppSecret '%v'", dc.AppSecret)
	}
	if dc.AccessToken == "" {
		return fmt.Errorf("Invalid AccessToken '%v'", dc.AccessToken)
	}
	if dc.BasePath == "" {
		return fmt.Errorf("BasePath can't be empty")
	}
	return nil
}

// Conversions

func DropboxConfigFromMap(m map[string]interface{}) (DropboxConfig, error) {
	var cfg DropboxConfig

	AppKey, ok := m["AppKey"].(string)
	if !ok {
		return cfg, fmt.Errorf("Invalid AppKey '%v'", m["AppKey"])
	}
	cfg.AppKey = AppKey

	AppSecret, ok := m["AppSecret"].(string)
	if !ok {
		return cfg, fmt.Errorf("Invalid AppSecret '%v'", m["AppSecret"])
	}
	cfg.AppSecret = AppSecret

	AccessToken, ok := m["AccessToken"].(string)
	if !ok {
		return cfg, fmt.Errorf("Invalid AccessToken '%v'", m["AccessToken"])
	}
	cfg.AccessToken = AccessToken

	BasePath, ok := m["BasePath"].(string)
	if !ok {
		return cfg, fmt.Errorf("Invalid BasePath '%v'", m["BasePath"])
	}
	cfg.BasePath = BasePath

	return cfg, nil
}

func DropboxConfigToMap(cfg DropboxConfig) map[string]interface{} {
	return map[string]interface{}{
		"AppKey":      cfg.AppKey,
		"AppSecret":   cfg.AppSecret,
		"AccessToken": cfg.AccessToken,
		"BasePath":    cfg.BasePath,
	}
}
