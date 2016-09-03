// Steven Phillips / elimisteve
// 2016.08.11

package backend

import "fmt"

const (
	TypeDropboxRemote = "dropbox"
	TypeFileSystem    = "filesystem"
	TypeWebserver     = "webserver"
)

func LoadBackend(backendPath, backendName string) (Backend, error) {
	if backendName == "" {
		backendName = "default"
	}

	conf, err := ReadConfig(backendPath, backendName)
	if err != nil {
		return nil, err
	}

	typ := conf.GetType()

	switch typ {
	case TypeDropboxRemote:
		return DropboxRemoteFromConfig(conf)
	case TypeFileSystem:
		return NewFileSystem(conf)
	case TypeWebserver:
		return WebserverFromConfig(conf)
	}

	return nil, fmt.Errorf("Unrecognized config `%v` of type `%v`",
		conf.Name, typ)
}
