// Steven Phillips / elimisteve
// 2016.08.30

package backend

import (
	"fmt"
	"strings"
)

func Create(bkType, bkName string, args []string) (Backend, error) {
	if bkName == "" {
		return nil, fmt.Errorf("Invalid Backend name `%v`", bkName)
	}
	if len(args) < 1 {
		return nil, fmt.Errorf("Need > %v additional arguments", len(args))
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
		if len(args) != 1 {
			return nil, fmt.Errorf("Filesystem Backend needs 1 arg, not %v",
				len(args))
		}

		conf := &Config{
			Name:     bkName,
			Type:     TypeFileSystem,
			Local:    true,
			DataPath: args[0],
		}

		return NewFileSystem(conf)

	case TypeWebserver:
		// Parse Sandstorm web key
		if len(args) == 1 {
			info := strings.SplitN(args[0], "#", 2)
			if len(info) != 2 {
				fmt.Errorf("Error parsing invalid Sandstorm web key `%s`",
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

	return nil, fmt.Errorf("Backend of type `%v` not recognized: %v", bkType)
}
