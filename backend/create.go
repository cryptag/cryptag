// Steven Phillips / elimisteve
// 2016.08.30

package backend

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/cryptag/cryptag"
)

var (
	ErrNilConfig = errors.New("Backend Config is nil")
)

// CreateFromConfig persists a new Backend Config to disk using cfg,
// then returns a new Backend based on cfg.
func CreateFromConfig(bkPath string, cfg *Config) (Backend, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}
	if bkPath == "" {
		bkPath = cryptag.BackendPath
	}

	if err := cfg.Canonicalize(); err != nil {
		return nil, err
	}

	// Make new in-memory Backend based on the *Config
	bk, err := New(cfg)
	if err != nil {
		return nil, err
	}

	// Config is OK; persist to disk
	if err := cfg.Save(bkPath); err != nil {
		return nil, err
	}

	return bk, nil
}

// New makes a Backend based on cfg. Does not persist new *Config to
// disk.
func New(cfg *Config) (Backend, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	bkMaker, err := GetMaker(cfg.GetType())
	if err != nil {
		return nil, err
	}

	return bkMaker(cfg)
}

// Create persists a new Backend Config to disk. DEPRECATED; use
// CreateFromConfig instead.
func Create(bkType, bkName string, args []string) (Backend, error) {
	if bkName == "" {
		return nil, fmt.Errorf("Invalid Backend name `%v`", bkName)
	}

	switch bkType {
	case TypeDropboxRemote:
		if len(args) != 4 {
			return nil, fmt.Errorf("Dropbox Backend needs 4 args, not %v",
				len(args))
		}
		cfg := DropboxConfig{
			AppKey:      args[0],
			AppSecret:   args[1],
			AccessToken: args[2],
			BasePath:    args[3],
		}
		return NewDropboxRemote(nil, bkName, cfg)

	case TypeFileSystem:
		if len(args) > 1 {
			return nil, fmt.Errorf("Filesystem Backend needs 0 or 1 args, not %v",
				len(args))
		}

		var dataPath string
		if len(args) == 1 {
			dataPath = args[0]
		} else {
			dataPath = path.Join(cryptag.LocalDataPath, "backends", bkName)
		}

		conf := &Config{
			Name:     bkName,
			Type:     TypeFileSystem,
			New:      true,
			Local:    true,
			DataPath: dataPath,
		}

		return NewFileSystem(conf)

	case TypeWebserver:
		// Parse Sandstorm web key
		if len(args) == 1 {
			info := strings.SplitN(args[0], "#", 2)
			if len(info) != 2 {
				return nil, fmt.Errorf("Error parsing invalid Sandstorm web key `%s`",
					args[0])
			}
			args = info
		}

		if len(args) != 2 {
			return nil, fmt.Errorf("Webserver Backend needs 2 args, not %v",
				len(args))
		}

		baseURL := args[0]
		authToken := args[1]

		return CreateWebserver(nil, bkName, baseURL, authToken)
	}

	return nil, fmt.Errorf("Backend type `%v` not recognized", bkType)
}
