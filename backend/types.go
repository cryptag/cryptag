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

	maker, err := GetMaker(typ)
	if err != nil {
		return nil, fmt.Errorf("Error getting Backend maker of type `%v`: %v",
			typ, err)
	}

	return maker(conf)
}
