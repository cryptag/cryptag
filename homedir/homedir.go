// Steve Phillips / elimisteve
// 2017.03.27

package homedir

import (
	"errors"
	"strings"

	gohomedir "github.com/mitchellh/go-homedir"
)

// Collapse turns paths of the form '/home/USER/rest/of/the/path' to
// '~/rest/of/the/path' provided that '/home/USER' is the current
// user's home directory.
func Collapse(path string) (collapsed string, err error) {
	if len(path) == 0 {
		return path, nil
	}

	// Already collapsed
	if path[0] == '~' {
		return path, nil
	}

	// If this isn't an absolute path ('/...' on Unix or 'C:\...' or
	// 'D:\...' etc on Windows), error out
	if path[0] != '/' && (len(path) > 2 && path[2] != '\\') {
		return "", errors.New("Cannot expand path that is neither absolute nor begins with '~'")
	}

	dir, err := gohomedir.Dir()
	if err != nil {
		return "", err
	}

	// Collapse
	if strings.HasPrefix(path, dir) {
		return "~" + path[len(dir):], nil
	}

	// Shouldn't collapse, so don't
	return path, nil
}
