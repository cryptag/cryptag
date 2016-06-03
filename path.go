// Steve Phillips / elimisteve
// 2015.11.14

package cryptag

import (
	"log"
	"os"
	"path"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
)

var (
	home, errHome = homedir.Dir()

	// TrustedBasePath is where private, non-shared data (e.g., crypto
	// keys) can be safely stored.
	TrustedBasePath = path.Join(home, ".cryptag")

	// BackendPath is the path to the directory where backend config
	// files are stored (e.g., "/home/myusername/.cryptag/backends").
	// This can be overridden with the BACKEND_PATH
	// environment variable (rarely useful; this exists so that all
	// CrypTag backend configs could be on a USB drive).
	BackendPath = path.Join(home, ".cryptag", "backends")

	// DefaultLocalDataPath is the default directory where local
	// backends will be told to store their data.  This directory will
	// contain 'rows' and 'tags' subdirectories.
	DefaultLocalDataPath = path.Join(home, ".cryptag")

	LocalDataPath = DefaultLocalDataPath
)

func init() {
	if errHome != nil {
		log.Printf("Error getting/setting home directory path: %v\n", errHome)
	}

	if p := os.Getenv("BACKEND_PATH"); p != "" {
		BackendPath = p
	}

	// Change LocalDataPath if apparently on Android
	if runtime.GOOS == "linux" && runtime.GOARCH == "arm" {
		androidStorage := "/storage/sdcard0"
		if _, err := os.Stat(androidStorage); os.IsNotExist(err) {
			LocalDataPath = androidStorage + "/.cryptag"
		}
	}

	// Change LocalDataPath if on Sandstorm (useful for servers)
	if os.Getenv("SANDSTORM") == "1" {
		LocalDataPath = "/var"
	}
}
