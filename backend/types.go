// Steven Phillips / elimisteve
// 2016.08.11

package backend

import (
	"errors"
	"fmt"
)

const (
	TypeDropboxRemote = "dropbox"
	TypeFileSystem    = "filesystem"
	TypeWebserver     = "webserver"
	TypeSandstorm     = "sandstorm" // Uses webserver + WebserverBackend code
)

var (
	ErrNoDefaultBackend = errors.New("backend: no default backend set;" +
		" specify a backend explicitly or set a default to use")
)

func LoadBackend(backendPath, backendName string) (Backend, error) {
	if backendName == "" {
		defaultExists, _ := IsDefaultBackendSet(backendPath)
		if !defaultExists {
			return nil, ErrNoDefaultBackend
		}
		backendName = "default"
	}

	conf, err := ReadConfig(backendPath, backendName)
	if err != nil {
		return nil, err
	}

	typ := conf.GetType()

	maker, err := GetMaker(typ)
	if err != nil {
		return nil, fmt.Errorf("Error getting Backend maker of type `%v`: %v",
			typ, err)
	}

	return maker(conf)
}
