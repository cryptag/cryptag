// Steve Phillips / elimisteve
// 2015.11.04

package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/types"
	"github.com/elimisteve/fun"
)

var (
	ErrConfigExists = errors.New("Backend config already exists")
)

type Config struct {
	Name     string
	Type     string // Should be one of: backend.Type*
	New      bool   `json:"-"`
	Key      *[32]byte
	Local    bool
	DataPath string // Used by backend.FileSystem, other local backends

	Custom map[string]interface{} `json:",omitempty"` // Used by Dropbox, Webserver, other backends
}

func (conf *Config) Save(backendsDir string) error {
	overwrite := false
	return conf.save(backendsDir, overwrite)
}

func (conf *Config) Update(backendsDir string) error {
	overwrite := true
	return conf.save(backendsDir, overwrite)
}

func (conf *Config) save(backendsDir string, overwrite bool) error {
	if err := os.MkdirAll(backendsDir, 0700); err != nil && os.IsExist(err) {
		return err
	}

	filename := path.Join(backendsDir, conf.Name+".json")

	if !overwrite {
		if _, err := os.Stat(filename); err == nil {
			log.Printf("Backend config already exists at %v; NOT overwriting",
				filename)
			return ErrConfigExists
		}
	}

	if err := conf.Canonicalize(); err != nil {
		return err
	}
	b, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	if overwrite {
		if err := conf.Backup(backendsDir); err != nil {
			return err
		}
	}

	if err = ioutil.WriteFile(filename, b, 0600); err != nil {
		return err
	}
	log.Printf("Saved backend config: %v\n", filename)

	return nil
}

func (conf *Config) Backup(backendsDir string) error {
	filename := path.Join(backendsDir, conf.Name+".json")

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Back up old config
	bkup := filename + "-" + cryptag.NowStr()

	if err = ioutil.WriteFile(bkup, b, 0600); err != nil {
		return err
	}

	log.Printf("Backed up %v to %v\n", filename, bkup)

	return nil
}

func (conf *Config) Canonicalize() error {
	if conf.Name == "" {
		return errors.New("Storage backend name cannot be empty")
	}
	conf.Name = strings.TrimSuffix(conf.Name, ".json")

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

	if conf.Local && conf.DataPath == "" {
		conf.DataPath = cryptag.LocalDataPath
	}
	conf.DataPath = strings.TrimRight(conf.DataPath, "/\\")

	return nil
}

//
// Convenience Functions
//

func ReadConfig(backendPath, backendName string) (*Config, error) {
	if backendPath == "" {
		backendPath = cryptag.BackendPath
	}
	if backendName == "" {
		return nil, errors.New("backendName cannot be empty")
	}
	backendName = strings.TrimSuffix(backendName, ".json")

	configFile := path.Join(backendPath, backendName+".json")

	if types.Debug {
		log.Printf("Loading backend config file `%v`\n", configFile)
	}

	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var conf Config
	err = json.Unmarshal(b, &conf)

	return &conf, err
}
