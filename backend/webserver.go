// Steve Phillips / elimisteve
// 2015.03.01

package backend

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elimisteve/cryptag/types"
	"github.com/elimisteve/fun"
)

var (
	HttpGetTimeout = 30 * time.Second
)

type WebserverBackend struct {
	serverBaseUrl string

	// cachedTags types.TagPairs

	// Used for encryption/decryption
	block cipher.Block
}

func NewWebserverBackend(aesKey []byte, serverBaseUrl string) (*WebserverBackend, error) {
	wb := &WebserverBackend{serverBaseUrl: serverBaseUrl}

	// Set wb.block
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	wb.block = block

	return wb, nil
}

func (wb *WebserverBackend) Block() cipher.Block {
	return wb.block
}

func (wb *WebserverBackend) AllTagPairs() (types.TagPairs, error) {
	return getTagsFrom(wb, wb.serverBaseUrl+"/tags")
}

func (wb *WebserverBackend) SaveRow(r *types.Row) (*types.Row, error) {
	// Populate row.{Encrypted,RandomTags} from
	// row.{decrypted,plainTags}
	row, err := PopulateRowBeforeSave(wb, r)
	if err != nil {
		return nil, err
	}

	rowBytes, err := json.Marshal(row)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling row: %v", err)
	}

	if types.Debug {
		log.Printf("POSTing row data: `%s`\n", rowBytes)
	}

	resp, err := http.Post(wb.serverBaseUrl, "application/json",
		bytes.NewReader(rowBytes))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got HTTP %d from server: `%s`", resp.StatusCode, body)
	}

	newRow, err := types.NewRowFromBytes(body)
	if err != nil {
		return nil, err
	}

	// Populated newRow.{decrypted,plainTags} from
	// newRow.{Encrypted,RandomTags}
	if err = PopulateRowAfterGet(wb, newRow); err != nil {
		return nil, err
	}

	return newRow, nil
}

func (wb *WebserverBackend) SaveTagPair(pair *types.TagPair) (*types.TagPair, error) {
	pairBytes, err := json.Marshal(pair)
	if err != nil {
		return nil, err
	}

	if types.Debug {
		log.Printf("POSTing tag pair data: `%s`\n", pairBytes)
	}

	resp, err := http.Post(wb.serverBaseUrl+"/tags", "application/json",
		bytes.NewReader(pairBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Error
	if resp.StatusCode != 200 {
		// Read server response to debug
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Got HTTP %d from server for data: `%s`",
			resp.StatusCode, body)
	}

	if types.Debug {
		log.Printf("New *TagPair created: `%#v`\n", pair)
	}

	return pair, nil
}

func (wb *WebserverBackend) TagPairsFromRandomTags(randtags []string) (types.TagPairs, error) {
	if len(randtags) == 0 {
		return nil, fmt.Errorf("Can't get 0 tags")
	}

	url := wb.serverBaseUrl + "/tags?tags=" + strings.Join(randtags, ",")
	return getTagsFrom(wb, url)
}

func (wb *WebserverBackend) RowsFromPlainTags(plaintags []string) (types.Rows, error) {
	randtags, err := randomFromPlain(wb, plaintags)
	if err != nil {
		return nil, fmt.Errorf("Error from RandomTagsFromPlain: %v", err)
	}
	if types.Debug {
		log.Printf("After randomTagsFromPlain: randtags == `%#v`\n", randtags)
	}

	fullURL := wb.serverBaseUrl + "?tags=" + strings.Join(randtags, ",")
	if types.Debug {
		log.Printf("fullURL == `%s`\n", fullURL)
	}

	return getRowsFrom(wb, fullURL)
}

//
// Helpers
//

// getRowsFrom fetches the encrypted rows from url, decrypts them, then
func getRowsFrom(backend Backend, url string) (types.Rows, error) {
	var rows types.Rows
	var err error

	if err = fun.FetchInto(url, HttpGetTimeout, &rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		if err = PopulateRowAfterGet(backend, row); err != nil {
			return nil, err
		}
	}

	return rows, nil
}

// getTagsFrom fetches the encrypted tag pairs at url, decrypts them,
// and unmarshals them into a TagPairs value
func getTagsFrom(backend Backend, url string) (types.TagPairs, error) {
	var pairs types.TagPairs
	var err error

	if err = fun.FetchInto(url, HttpGetTimeout, &pairs); err != nil {
		return nil, fmt.Errorf("Error fetching pairs: %v", err)
	}

	for _, pair := range pairs {
		if err = pair.Decrypt(backend.Block()); err != nil {
			return nil, err
		}
	}

	return pairs, nil
}
