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

func (wb *WebserverBackend) Key() *[32]byte {
	return wb.key
}

func (wb *WebserverBackend) AllTagPairs() (types.TagPairs, error) {
	pairs, err := getTagsFromUrl(wb.key, wb.tagsUrl, wb.authToken)
	if err != nil {
		return nil, err
	}
	if types.Debug {
		log.Printf("AllTagPairs: returning %d pairs\n", len(pairs))
	}
	return pairs, nil
}

func (wb *WebserverBackend) SaveRow(row *types.Row) error {
	// Populate row.{Encrypted,RandomTags} from
	// row.{decrypted,plainTags}
	err := PopulateRowBeforeSave(wb, row)
	if err != nil {
		return fmt.Errorf("Error populating row before save: %v", err)
	}

	rowBytes, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("Error marshaling row: %v", err)
	}

	if types.Debug {
		log.Printf("POSTing row data: `%s`\n\n", rowBytes)
	}

	resp, err := post(wb.rowsUrl, rowBytes, wb.authToken)
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

	// // Populated newRow.{decrypted,plainTags} from
	// // newRow.{Encrypted,RandomTags}
	// if err = PopulateRowAfterGet(wb, newRow); err != nil {
	// 	return err
	// }

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

	resp, err := post(wb.tagsUrl, pairBytes, wb.authToken)
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
	return getTagsFromUrl(wb.key, url, wb.authToken)
}

func (wb *WebserverBackend) ListRows(randtags cryptag.RandomTags) (types.Rows, error) {
	return nil, fmt.Errorf("WebserverBackend.ListRows: NOT IMPLEMENTED")
}

func (wb *WebserverBackend) RowsFromRandomTags(randtags cryptag.RandomTags) (types.Rows, error) {
	fullURL := wb.rowsUrl + "?tags=" + strings.Join(randtags, ",")
	if types.Debug {
		log.Printf("fullURL == `%s`\n", fullURL)
	}

	rows, err := getRowsFromUrl(fullURL, wb.authToken)
	if err != nil {
		return nil, fmt.Errorf("Error from getRowsFromUrl: %v", err)
	}
	return rows, nil
}

func (wb *WebserverBackend) DeleteRows(randtags cryptag.RandomTags) error {
	return errors.New("WebserverBackend.DeleteRows NOT IMPLEMENTED")
}

//
// Helpers
//

// getRowsFromUrl fetches the encrypted rows from url. Does not
// decrypt and populate them.
func getRowsFromUrl(url, authToken string) (types.Rows, error) {
	var rows types.Rows

	err := getInto(url, authToken, &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// getTagsFromUrl fetches the encrypted tag pairs at url, decrypts them,
// and unmarshals them into a TagPairs value
func getTagsFromUrl(key *[32]byte, url, authToken string) (types.TagPairs, error) {
	if key == nil {
		return nil, cryptag.ErrNilKey
	}

	var pairs types.TagPairs
	var err error

	if types.Debug {
		log.Printf("getTagsFromUrl: Getting tags from URL `%v`\n", url)
	}

	if err = getInto(url, authToken, &pairs); err != nil {
		return nil, fmt.Errorf("Error fetching pairs: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(pairs))

	for _, pair := range pairs {
		go func(pair *types.TagPair) {
			// TODO: Return first error
			if err = pair.Decrypt(key); err != nil {
				log.Printf("Error from pair.Decrypt: %v", err)
			}
			wg.Done()
		}(pair)
	}

	wg.Wait()

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
