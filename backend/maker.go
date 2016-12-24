// Steve Phillips / elimisteve
// 2016.12.21

package backend

import (
	"errors"
	"sync"
)

var (
	ErrMakerNotFound = errors.New("Backend maker func not found")
)

// A maker is a function that creates a Backend from a Config
type maker func(*Config) (Backend, error)

// A makerMap stores a map of maker funcs that can be updated in a
// thread-safe way
type makerMap struct {
	mu sync.RWMutex
	m  map[string]maker
}

var makers = makerMap{
	m: map[string]maker{
		TypeDropboxRemote: func(cfg *Config) (Backend, error) {
			return DropboxRemoteFromConfig(cfg)
		},
		TypeFileSystem: func(cfg *Config) (Backend, error) {
			return NewFileSystem(cfg)
		},
		TypeWebserver: func(cfg *Config) (Backend, error) {
			return WebserverFromConfig(cfg)
		},
	},
}

// GetMaker returns a Backend maker function that will make a Backend
// of the specified type.  If no such Backend has been registered,
// ErrMakerNotFound is returned.
func GetMaker(bkType string) (maker, error) {
	makers.mu.RLock()
	defer makers.mu.RUnlock()

	f, exists := makers.m[bkType]
	if !exists {
		// return nil, fmt.Errorf("Backend type `%v` not recognized", cfg.Type)
		return nil, ErrMakerNotFound
	}

	return f, nil
}

// RegisterMaker registers a new Backend maker function by type.
// Suitable for placing in an init() function in the file where a new
// Backend type has been defined.
func RegisterMaker(bkType string, f maker) error {
	makers.mu.Lock()
	defer makers.mu.Unlock()

	// TODO: Return error if bkType already in cm? Probably not...

	makers.m[bkType] = f
	return nil
}
