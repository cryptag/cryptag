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

func RowTagWithPrefix(r *Row, prefix string) string {
	for _, t := range r.PlainTags() {
		if strings.HasPrefix(t, prefix) {
			return strings.TrimPrefix(t, prefix)
		}
	}
	return ""
}

func SaveRowAsFile(r *Row, dir string) (filepath string, err error) {
	f := RowTagWithPrefix(r, "filename:")
	if f == "" {
		f = RowTagWithPrefix(r, "id:")
		if f == "" {
			f = fmt.Sprintf("%v", time.Now().Unix())
		}
		f += "." + RowTagWithPrefix(r, "type:")
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
