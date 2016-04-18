// Steven Phillips / elimisteve
// 2016.01.11

package types

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/elimisteve/cryptag"
)

// RowTagWithPrefix iterates over r's plaintags to grab the data
// stored in said tags, preferring the prefixes in the order they are
// passed in.
//
// Example use: RowTagWithPrefix(r, "filename:", "id:"), which will
// try to return r's filename, then try to return its ID, and if both
// fail will return empty string.
func RowTagWithPrefix(r *Row, prefixes ...string) string {
	for _, prefix := range prefixes {
		for _, tag := range r.PlainTags() {
			if strings.HasPrefix(tag, prefix) {
				return strings.TrimPrefix(tag, prefix)
			}
		}
	}
	return ""
}

// SaveRowAsFile saves the given *Row r to the given directory dir,
// introspecting r to determine a reasonable filename (either the
// filename if r contains a file, or r's unique ID).  If dir is empty,
// r is stored in the "decrypted" directory within
// cryptag.TrustedBasePath ($HOME/.cryptag by default).
func SaveRowAsFile(r *Row, dir string) (filepath string, err error) {
	f := RowTagWithPrefix(r, "filename:", "id:")
	if f == "" {
		f = fmt.Sprintf("%d", cryptag.Now().Unix())
	}

	if dir == "" {
		dir = path.Join(cryptag.TrustedBasePath, "decrypted")
		if err := os.MkdirAll(dir, 0700); err != nil {
			log.Printf("Error creating directory `%v`: %v", dir, err)
		}
	}

	filepath = path.Join(dir, f)

	if err := ioutil.WriteFile(filepath, r.Decrypted(), 0644); err != nil {
		return "", err
	}

	return filepath, nil
}
