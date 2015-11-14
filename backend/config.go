// Steve Phillips / elimisteve
// 2015.11.04

package backend

import (
	"os"
	"path"
	"strings"

	"github.com/elimisteve/cryptag"
)

type Config struct {
	Key             *[32]byte
	Local           bool
	CryptagBasePath string // Used by backend.FileSystem, other local backends

	// BaseURL  string // Used by backend.Webserver, other remote backends
}

func (conf *Config) Canonicalize() error {
	// Create new key
	if conf.Key == nil {
		log.Printf("Generating new encryption key for backend `%s`...",
			conf.Name)
		key, err := cryptag.RandomKey()
		if err != nil {
			return err
		}
		conf.Key = key
	}

	if conf.Local && conf.CryptagBasePath == "" {
		conf.CryptagBasePath = path.Join(os.Getenv("HOME"), ".cryptag")
	}
	conf.CryptagBasePath = strings.TrimRight(conf.CryptagBasePath, "/\\")

	return nil
}
