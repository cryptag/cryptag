// Steve Phillips / elimisteve
// 2015.02.24

package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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
	row := &Row{decrypted: decrypted, plainTags: plainTags}
	log.Printf("New Row created: `%#v`\n", row)
	return row
}

func (row *Row) Save() error {
	// Fetch all tags.  For each element of row.plainTags that doesn't
	// match an existing tag, call CreateTag().  Encrypt row.decrypted
	// and store it in row.Encrypted.  POST to server.
	allTagPairs, _, err := CreateTagsFromPlain(row.plainTags)
	if err != nil {
		return fmt.Errorf("Error from CreateNewTagsFromPlain: %v", err)
	}

	// Set row.Encrypted
	encData, err := crypt.AESEncryptBytes(Block, row.decrypted)
	if err != nil {
		return fmt.Errorf("Error encrypting data: %v", err)
	}
	row.Encrypted = encData

	// Set row.RandomTags
	for _, pair := range allTagPairs {
		if row.HasPlainTag(pair.plain) {
			row.RandomTags = append(row.RandomTags, pair.Random)
		}
	}

	log.Printf("row, just before POST: `%#v`\n", row)

	// POST to server
	if err := row.post(); err != nil {
		return fmt.Errorf("Error POSTing row to server: %v", err)
	}

	return nil
}

func (row *Row) post() error {
	rowBytes, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("Error marshaling row: %v", err)
	}

	resp, err := http.Post(SERVER_BASE_URL, "application/json",
		bytes.NewReader(rowBytes))

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Got HTTP %d from server: `%s`", resp.StatusCode, body)
	}

	return nil
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

// decryptData sets row.decrypted based upon row.Encrypted
func (row *Row) decryptData() error {
	if len(row.Encrypted) == 0 {
		return fmt.Errorf("no data to decrypt")
	}

	dec, err := crypt.AESDecryptBytes(Block, row.decrypted)
	if err != nil {
		return fmt.Errorf("Error decrypting: %v", err)
	}

	row.decrypted = dec

	return nil
}

// setPlainTags uses row.RandomTags and retrieved TagPairs to set
// row.plainTags
func (row *Row) setPlainTags() error {
	pairs, err := GetTagPairsFromRandom(row.RandomTags...)
	if err != nil {
		return err
	}

	row.plainTags = pairs.AllPlain()
	log.Printf("row.plainTags set to `%v`\n", row.plainTags)

	return nil
}
