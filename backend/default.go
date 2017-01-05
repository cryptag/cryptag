// Steven Phillips / elimisteve
// 2016.08.11

package backend

import (
	"fmt"
	"os"
)

// SetDefaultBackend sets the default Backend to newDefault by
// creating a symlink from (newDefault).json to default.json
func SetDefaultBackend(backendPath, newDefault string) error {
	conf := ConfigPathFromName(backendPath, newDefault)
	defaultConf := ConfigPathFromName(backendPath, "default")

	defaultInfo, err := os.Lstat(defaultConf)
	if os.IsNotExist(err) {
		// If default.json doesn't exist, it's safe to create it as a
		// symlink
		return os.Symlink(conf, defaultConf)
	}
	if err != nil {
		return err
	}

	// default.json exists. If it's not a symlink, error out.

	if defaultInfo.Mode()&os.ModeSymlink != os.ModeSymlink {
		return fmt.Errorf("%s is not a symlink. Don't name"+
			" your backends 'default'!", defaultConf)
	}

	// default.json exists and is a symlink. Delete it and re-symlink.

	if err = os.Remove(defaultConf); err != nil {
		return err
	}

	return os.Symlink(conf, defaultConf)
}

// Checks if a default backend is set.  Returns true if so, false
// otherwise.
func IsDefaultBackendSet(backendPath string) (bool, error) {
	defaultConf := ConfigPathFromName(backendPath, "default")

	defaultInfo, err := os.Lstat(defaultConf)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// default.json exists. If it's not a symlink, error out.

	if defaultInfo.Mode()&os.ModeSymlink != os.ModeSymlink {
		return false, fmt.Errorf("%s is not a symlink. Don't name"+
			" your backends 'default'!", defaultConf)
	}

	// default.json exists and is a symlink
	return true, nil
}
