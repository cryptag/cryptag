package rowutil

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/types"
)

// SaveAsFile saves the given *Row r to the given directory dir,
// introspecting r to determine a reasonable filename (either the
// filename if r contains a file, or r's unique ID).  If dir is empty,
// r is stored in the "decrypted" directory within
// cryptag.TrustedBasePath ($HOME/.cryptag by default).
func SaveAsFile(r *types.Row, dir string) (filepath string, err error) {
	f := TagWithPrefixStripped(r, "filename:", "id:")
	if f == "" {
		log.Printf("Warning: row doesn't have an id:... tag!\n")
		f = fmt.Sprintf("%d", cryptag.Now().Unix())
	}

	if dir == "" {
		dir = path.Join(cryptag.TrustedBasePath, "decrypted")
	}

	// Create destination dir
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("Error creating directory `%v`: %v", dir, err)
	}

	filepath = path.Join(dir, f)

	if err := ioutil.WriteFile(filepath, r.Decrypted(), 0600); err != nil {
		return "", err
	}

	return filepath, nil
}
