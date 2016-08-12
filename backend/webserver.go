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
	"sync"
	"time"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/tor"
	"github.com/elimisteve/cryptag/types"
)

var (
	HttpGetTimeout = 300 * time.Second
)

type WebserverBackend struct {
	serverName    string
	serverBaseUrl string
	rowsUrl       string
	tagsUrl       string

	client *http.Client
	useTor bool

	authToken string

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
		client:        &http.Client{},
	}

	return ws, nil
}

// CreateWebserver is a high-level function that creates a new
// WebserverBackend and saves its config to disk. (Useful for config
// init and more.)
func CreateWebserver(key []byte, backendName, baseURL, authToken string) (*WebserverBackend, error) {
	db, err := NewWebserverBackend(key, backendName, baseURL, authToken)
	if err != nil {
		return nil, fmt.Errorf("Error from NewWebserverBackend: %v", err)
	}

	cfg, err := db.ToConfig()
	if err != nil {
		return db, fmt.Errorf("Error converting WebserverBackend to Config: %v",
			err)
	}

	err = cfg.Save(cryptag.BackendPath)
	if err != nil {
		return db, fmt.Errorf("Error saving backend config to disk: %v", err)
	}

	return db, nil
}

func LoadWebserverBackend(backendPath, backendName string) (*WebserverBackend, error) {
	if backendName == "" {
		backendName = "webserver"
	}

	conf, err := ReadConfig(backendPath, backendName)
	if err != nil {
		return nil, err
	}

	return WebserverFromConfig(conf)
}

func WebserverFromConfig(conf *Config) (*WebserverBackend, error) {
	if conf.Key == nil {
		return nil, fmt.Errorf("Key cannot be empty!")
	}

	webConf, err := WebserverConfigFromMap(conf.Custom)
	if err != nil {
		return nil, err
	}

	return NewWebserverBackend((*conf.Key)[:], conf.Name, webConf.BaseURL,
		webConf.AuthToken)
}

func (wb *WebserverBackend) ToConfig() (*Config, error) {
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

// SetHTTPClient sets the underlying HTTP client used. Useful for
// using a custom client that does proxied requests, perhaps through
// Tor.
func (wb *WebserverBackend) SetHTTPClient(client *http.Client) {
	wb.client = client
}

// UseTor sets wb's HTTP client to one that uses Tor and records that
// Tor should be used.
func (wb *WebserverBackend) UseTor() error {
	client, err := tor.NewClient()
	if err != nil {
		return err
	}

	wb.SetHTTPClient(client)
	wb.useTor = true

	if types.Debug {
		log.Println("*WebserverBackend to do HTTP calls over Tor")
	}

	return nil
}

func (wb *WebserverBackend) Name() string {
	return wb.serverName
}

func (wb *WebserverBackend) Key() *[32]byte {
	return wb.key
}

func (wb *WebserverBackend) AllTagPairs(oldPairs types.TagPairs) (types.TagPairs, error) {
	pairs, err := wb.getTagsFromUrl(wb.tagsUrl)
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

func (wb *WebserverBackend) SaveRow(row *types.Row) error {
	if len(row.Encrypted) == 0 || len(row.RandomTags) == 0 || row.Nonce == nil || *row.Nonce == [24]byte{} {
		return errors.New("Invalid row; requires Encrypted, RandomTags, Nonce fields")
	}

	rowBytes, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("Error marshaling row: %v", err)
	}

	if types.Debug {
		log.Printf("POSTing row data: `%s`\n\n", rowBytes)
	}

	resp, err := wb.post(wb.rowsUrl, rowBytes)
	if err != nil {
		return fmt.Errorf("Error POSTing row to URL %s: %v", wb.rowsUrl, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading server response body: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Got HTTP %d from server: `%s`", resp.StatusCode, body)
	}

	return nil
}

func (wb *WebserverBackend) SaveTagPair(pair *types.TagPair) error {
	pairBytes, err := json.Marshal(pair)
	if err != nil {
		return err
	}

	if types.Debug {
		log.Printf("POSTing tag pair data: `%s`\n\n", pairBytes)
	}

	resp, err := wb.post(wb.tagsUrl, pairBytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Error
	if resp.StatusCode != 200 {
		// Read server response to debug
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("Got HTTP %d from server for data: `%s`",
			resp.StatusCode, body)
	}

	if types.Debug {
		log.Printf("New *TagPair created: `%#v`\n", pair)
	}

	return nil
}

func (wb *WebserverBackend) TagPairsFromRandomTags(randtags cryptag.RandomTags) (types.TagPairs, error) {
	if len(randtags) == 0 {
		return nil, fmt.Errorf("Can't get 0 tags")
	}

	url := wb.tagsUrl + "?tags=" + strings.Join(randtags, ",")
	return wb.getTagsFromUrl(url)
}

func (wb *WebserverBackend) ListRows(randtags cryptag.RandomTags) (types.Rows, error) {
	fullURL := wb.rowsUrl + "/list?tags=" + strings.Join(randtags, ",")
	return wb.getRowsFromUrl(fullURL)
}

func (wb *WebserverBackend) RowsFromRandomTags(randtags cryptag.RandomTags) (types.Rows, error) {
	fullURL := wb.rowsUrl + "?tags=" + strings.Join(randtags, ",")
	return wb.getRowsFromUrl(fullURL)
}

func (wb *WebserverBackend) DeleteRows(randtags cryptag.RandomTags) error {
	fullURL := wb.rowsUrl + "/delete?tags=" + strings.Join(randtags, ",")
	resp, err := wb.get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Error deleting rows; got status code %d and body `%s`",
			resp.StatusCode, body)
	}

	return nil
}

//
// Helper Methods
//

// getRowsFromUrl fetches the encrypted rows from url. Does not
// decrypt and populate them.
func (wb *WebserverBackend) getRowsFromUrl(url string) (types.Rows, error) {
	var rows types.Rows

	if types.Debug {
		log.Printf("getRowsFromUrl: Getting rows from URL `%v`\n", url)
	}

	err := wb.getInto(url, &rows)
	if err != nil {
		return nil, err
	}

	if types.Debug {
		log.Printf("getRowsFromUrl: returning %d Rows\n", len(rows))
	}

	return rows, nil
}

// getTagsFromUrl fetches the encrypted tag pairs at url, decrypts
// them, and unmarshals them into a TagPairs value
func (wb *WebserverBackend) getTagsFromUrl(url string) (types.TagPairs, error) {
	var pairs types.TagPairs
	var err error

	if types.Debug {
		log.Printf("getTagsFromUrl: Getting tags from URL `%v`\n", url)
	}

	if err = wb.getInto(url, &pairs); err != nil {
		return nil, fmt.Errorf("Error fetching pairs: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(pairs))

	for _, pair := range pairs {
		go func(pair *types.TagPair) {
			// TODO: Return first error
			if err = pair.Decrypt(wb.key); err != nil {
				log.Printf("Error from pair.Decrypt: %v", err)
			}
			wg.Done()
		}(pair)
	}

	wg.Wait()

	if types.Debug {
		log.Printf("getTagsFromUrl: returning %d TagPairs\n", len(pairs))
	}

	return pairs, nil
}

func (wb *WebserverBackend) get(url string) (*http.Response, error) {
	reqBuilder := http.NewRequest
	if wb.useTor {
		reqBuilder = tor.NewRequest
	}

	req, err := reqBuilder("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+wb.authToken)

	return wb.client.Do(req)
}

func (wb *WebserverBackend) getInto(url string, strct interface{}) error {
	resp, err := wb.get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if 400 <= resp.StatusCode && resp.StatusCode <= 599 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d from %s; response: %s", resp.StatusCode,
			url, body)
	}

	return readInto(resp.Body, strct)
}

func (wb *WebserverBackend) post(url string, data []byte) (*http.Response, error) {
	reqBuilder := http.NewRequest
	if wb.useTor {
		reqBuilder = tor.NewRequest
	}

	req, err := reqBuilder("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Error creating POST request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+wb.authToken)

	return wb.client.Do(req)
}

//
// Helpers
//

func readInto(r io.Reader, strct interface{}) error {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, strct)
	if err != nil {
		return fmt.Errorf("Error reading body `%s` into Go type: %v", body, err)
	}

	return nil
}
