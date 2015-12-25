// Steve Phillips / elimisteve
// 2015.11.14

package cryptag

import (
	"os"
	"path"
)

var (
	// BackendPath is the path to the directory where backend config
	// files are stored (e.g., "/home/myusername/.cryptag/backends").
	// This can be overridden with the CRYPTAG_BACKEND_PATH
	// environment variable (rarely useful; this exists so that all
	// CrypTag backend configs could be on a USB drive).
	BackendPath = path.Join(os.Getenv("HOME"), ".cryptag", "backends")

	// DefaultLocalDataPath is the default directory where local
	// backends will be told to store their data.  This directory will
	// contain 'rows' and 'tags' subdirectories.
	DefaultLocalDataPath = path.Join(os.Getenv("HOME"), ".cryptag")

	LocalDataPath = DefaultLocalDataPath
)

func init() {
	if p := os.Getenv("CRYPTAG_BACKEND_PATH"); p != "" {
		BackendPath = p
	}
}
