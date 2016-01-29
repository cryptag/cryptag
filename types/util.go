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
	"time"

	"github.com/elimisteve/cryptag"
)

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

func SaveRowAsFile(r *Row, dir string) (filepath string, err error) {
	f := RowTagWithPrefix(r, "filename:", "id:")
	if f == "" {
		f = fmt.Sprintf("%d", time.Now().Unix())
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
