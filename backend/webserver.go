// Steve Phillips / elimisteve
// 2015.03.01

package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/types"
)

var (
	HttpGetTimeout = 30 * time.Second
)

type WebserverBackend struct {
	serverName    string
	serverBaseUrl string
	rowsUrl       string
	tagsUrl       string

	// cachedTags types.TagPairs

	authToken string

	// Used for encryption/decryption
	key *[32]byte
}

func NewWebserverBackend(key []byte, serverName, serverBaseUrl, authToken string) (*WebserverBackend, error) {
	if serverBaseUrl == "" {
		return nil, fmt.Errorf("Invalid serverBaseUrl `%s`", serverBaseUrl)
	}
	serverBaseUrl = strings.TrimRight(serverBaseUrl, "/")

	if len(key) == 0 {
		good, err := cryptag.RandomKey()
		if err != nil {
			return nil, err
		}
		// TODO: Shouldn't have to do this...
		key = (*good)[:]
	}

	goodKey, err := cryptag.ConvertKey(key)
	if err != nil {
		return nil, err
	}

	ws := &WebserverBackend{
		key:           goodKey,
		serverName:    serverName,
		serverBaseUrl: serverBaseUrl,
		rowsUrl:       serverBaseUrl + "/rows",
		tagsUrl:       serverBaseUrl + "/tags",
		authToken:     authToken,
	}

	return ws, nil
}

func LoadWebserverBackend(backendPath, backendName string) (*WebserverBackend, error) {
	if backendPath == "" {
		backendPath = cryptag.BackendPath
	}
	if backendName == "" {
		backendName = "webserver"
	}
	backendName = strings.TrimSuffix(backendName, ".json")

	configFile := path.Join(backendPath, backendName+".json")

	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// Config exists

	var conf Config
	if err := json.Unmarshal(b, &conf); err != nil {
		return nil, err
	}

	if conf.Key == nil {
		return nil, fmt.Errorf("Key cannot be empty!")
	}

	webConf, err := WebserverConfigFromMap(conf.Custom)
	if err != nil {
		return nil, err
	}

	return NewWebserverBackend((*conf.Key)[:], backendName, webConf.BaseURL,
		webConf.AuthToken)
}

func (wb *WebserverBackend) Config() (*Config, error) {
	if wb.key == nil {
		return nil, cryptag.ErrNilKey
	}
	c := Config{
		Name: wb.serverName,
		Key:  wb.key,
		Custom: map[string]interface{}{
			"AuthToken": wb.authToken,
			"BaseURL":   wb.serverBaseUrl,
		},
	}
	return &c, nil
}

func (wb *WebserverBackend) Encrypt(plain []byte, nonce *[24]byte) (cipher []byte, err error) {
	return cryptag.Encrypt(plain, nonce, wb.key)
}

func (wb *WebserverBackend) Decrypt(cipher []byte, nonce *[24]byte) (plain []byte, err error) {
	return cryptag.Decrypt(cipher, nonce, wb.key)
}

func (wb *WebserverBackend) AllTagPairs() (types.TagPairs, error) {
	return getTagsFromUrl(wb, wb.tagsUrl, wb.authToken)
}

func (wb *WebserverBackend) SaveRow(r *types.Row) (*types.Row, error) {
	// Populate row.{Encrypted,RandomTags} from
	// row.{decrypted,plainTags}
	row, err := PopulateRowBeforeSave(wb, r)
	if err != nil {
		return nil, fmt.Errorf("Error populating row before save: %v", err)
	}

	rowBytes, err := json.Marshal(row)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling row: %v", err)
	}

	if types.Debug {
		log.Printf("POSTing row data: `%s`\n\n", rowBytes)
	}

	resp, err := post(wb.rowsUrl, rowBytes, wb.authToken)
	if err != nil {
		return nil, fmt.Errorf("Error POSTing row to URL %s: %v", wb.rowsUrl, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading server response body: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got HTTP %d from server: `%s`", resp.StatusCode, body)
	}

	newRow, err := types.NewRowFromBytes(body)
	if err != nil {
		return nil, fmt.Errorf("Error creating new row from server response: %v", err)
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
		log.Printf("POSTing tag pair data: `%s`\n\n", pairBytes)
	}

	resp, err := post(wb.tagsUrl, pairBytes, wb.authToken)
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

	url := wb.tagsUrl + "?tags=" + strings.Join(randtags, ",")
	return getTagsFromUrl(wb, url, wb.authToken)
}

func (wb *WebserverBackend) ListRows(plaintags []string) (types.Rows, error) {
	return nil, fmt.Errorf("WebserverBackend.ListRows: NOT IMPLEMENTED")
}

func (wb *WebserverBackend) RowsFromPlainTags(plaintags []string) (types.Rows, error) {
	randtags, err := randomFromPlain(wb, plaintags)
	if err != nil {
		return nil, fmt.Errorf("Error from RandomTagsFromPlain: %v", err)
	}
	if types.Debug {
		log.Printf("After randomTagsFromPlain: randtags == `%#v`\n", randtags)
	}

	fullURL := wb.rowsUrl + "?tags=" + strings.Join(randtags, ",")
	if types.Debug {
		log.Printf("fullURL == `%s`\n", fullURL)
	}

	rows, err := getRowsFromUrl(wb, fullURL, wb.authToken)
	if err != nil {
		return nil, fmt.Errorf("Error from getRowsFromUrl: %v", err)
	}
	return rows, nil
}

func (wb *WebserverBackend) DeleteRows(randTags []string) error {
	return errors.New("WebserverBackend.DeleteRows NOT IMPLEMENTED")
}

//
// Helpers
//

// getRowsFromUrl fetches the encrypted rows from url, decrypts them, then
func getRowsFromUrl(backend Backend, url, authToken string) (types.Rows, error) {
	var rows types.Rows

	err := getInto(url, authToken, &rows)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if err = PopulateRowAfterGet(backend, row); err != nil {
			return nil, fmt.Errorf("Error from PopulateRowAfterGet: %v", err)
		}
	}

	return rows, nil
}

// getTagsFromUrl fetches the encrypted tag pairs at url, decrypts them,
// and unmarshals them into a TagPairs value
func getTagsFromUrl(backend Backend, url, authToken string) (types.TagPairs, error) {
	var pairs types.TagPairs
	var err error

	if err = getInto(url, authToken, &pairs); err != nil {
		return nil, fmt.Errorf("Error fetching pairs: %v", err)
	}

	for _, pair := range pairs {
		if err = pair.Decrypt(backend.Decrypt); err != nil {
			return nil, fmt.Errorf("Error from pair.Decrypt: %v", err)
		}
	}

	return pairs, nil
}

func get(url, authToken string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+authToken)

	client := &http.Client{Timeout: HttpGetTimeout}

	return client.Do(req)
}

func getInto(url, authToken string, strct interface{}) error {
	resp, err := get(url, authToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return readInto(resp.Body, strct)
}

func readInto(r io.Reader, strct interface{}) error {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, strct)
}

func post(url string, data []byte, authToken string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Error creating POST request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+authToken)

	client := &http.Client{Timeout: HttpGetTimeout}

	return client.Do(req)
}

func postInto(url string, data []byte, authToken string, strct interface{}) error {
	resp, err := post(url, data, authToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return readInto(resp.Body, strct)
}
