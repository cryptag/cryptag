// Steve Phillips / elimisteve
// 2016.12.05

package trusted

import "github.com/cryptag/cryptag/backend"

// Config represents a trusted backend.Config file. Trivial other
// differences aside (namely the New and Path fields), it contains all
// the same fields as a backend.Config, except the 'Key' field is
// excluded here.
type Config struct {
	Name     string
	Type     string // Should be one of: backend.Type*
	Local    bool
	DataPath string // Used by backend.FileSystem, other local backends

	Custom map[string]interface{} `json:",omitempty"` // Used by Dropbox, Webserver, other backends

	Path string
}

type Configs []*Config

func FromConfigs(configs []*backend.Config) Configs {
	out := make(Configs, 0, len(configs))
	for _, in := range configs {
		out = append(out, FromConfig(in))
	}
	return out
}

func FromConfig(config *backend.Config) *Config {
	return &Config{
		Name:     config.Name,
		Type:     config.GetType(),
		Local:    config.Local,
		DataPath: config.DataPath,
		Custom:   config.Custom,
		Path:     config.GetPath(),
	}
}

func ToConfig(tcfg *Config) *backend.Config {
	return &backend.Config{
		Name:     tcfg.Name,
		Type:     tcfg.Type,
		Local:    tcfg.Local,
		DataPath: tcfg.DataPath,
		Custom:   tcfg.Custom,
	}

}
