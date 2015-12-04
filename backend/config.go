// Steve Phillips / elimisteve
// 2015.11.04

package backend

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/fun"
)

type Config struct {
	Name            string
	New             bool `json:"-"`
	Key             *[32]byte
	Local           bool
	BackendBasePath string // Used by backend.FileSystem, other local backends

	// BaseURL  string // Used by backend.Webserver, other remote backends
}

func (conf *Config) Canonicalize() error {
	if conf.Name == "" {
		return errors.New("Storage backend name cannot be empty")
	}
	if fun.ContainsAnyStrings(conf.Name, " ", "\t", "\r", "\n") {
		return fmt.Errorf("Storage backend name `%s` contains one or"+
			" more whitespace characters, shouldn't", conf.Name)
	}

	if conf.Key == nil {
		log.Printf("Generating new encryption key for backend `%s`...",
			conf.Name)
		key, err := cryptag.RandomKey()
		if err != nil {
			return err
		}
		conf.Key = key
	}

	if conf.Local && conf.BackendBasePath == "" {
		conf.BackendBasePath = cryptag.Path
	}
	conf.BackendBasePath = strings.TrimRight(conf.BackendBasePath, "/\\")

	return nil
}
