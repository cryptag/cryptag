// Steve Phillips / elimisteve
// 2015.02.24

package types

import (
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elimisteve/clipboard"
	"github.com/elimisteve/fun"
	"github.com/thecloakproject/utils/crypt"
)

type Row struct {
	// Populated by server
	Encrypted  []byte   `json:"data"`
	RandomTags []string `json:"tags"`

	// Populated locally
	decrypted []byte
	plainTags []string
}

func NewRow(decrypted []byte, plainTags []string) *Row {
	return &Row{decrypted: decrypted, plainTags: plainTags}
}

func NewRowFromBytes(body []byte) (*Row, error) {
	row := &Row{}
	if err := json.Unmarshal(body, row); err != nil {
		return nil, fmt.Errorf("Error creating new row: `%v`. Input: `%s`", err,
			body)
	}
	if Debug {
		log.Printf("Created new Row `%#v` from bytes: `%s`\n", row, body)
	}
	return row, nil
}

func (row *Row) Decrypted() []byte {
	return row.decrypted
}

func (row *Row) PlainTags() []string {
	return row.plainTags
}

func (row *Row) ToClipboard() error {
	dec := row.decrypted
	if Debug {
		log.Printf("Writing this to clipboard: `%s`\n", dec)
	}
	return clipboard.WriteAll(dec)
}

func (row *Row) HasRandomTag(randtag string) bool {
	return fun.SliceContains(row.RandomTags, randtag)
}

func (row *Row) HasPlainTag(plain string) bool {
	return fun.SliceContains(row.plainTags, plain)
}

func (row *Row) Format() string {
	return fmt.Sprintf("%s\t%s\n", row.decrypted, strings.Join(row.plainTags, "  "))
}

// DecryptData sets row.decrypted based upon row.Encrypted
func (row *Row) Decrypt(block cipher.Block) error {
	if len(row.Encrypted) == 0 {
		return fmt.Errorf("no data to decrypt")
	}

	dec, err := crypt.AESDecryptBytes(block, row.Encrypted)
	if err != nil {
		return fmt.Errorf("Error decrypting: %v", err)
	}

	row.decrypted = dec

	return nil
}

// SetPlainTags uses row.RandomTags and retrieved TagPairs to set
// row.plainTags
func (row *Row) SetPlainTags(getPairsFromRandom func(plain []string) (TagPairs, error)) error {
	pairs, err := getPairsFromRandom(row.RandomTags)
	if err != nil {
		return err
	}

	row.plainTags = pairs.AllPlain()
	if Debug {
		log.Printf("row.plainTags set to `%#v`\n", row.plainTags)
	}

	return nil
}
