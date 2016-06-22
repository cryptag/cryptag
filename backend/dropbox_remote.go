// Steven Phillips / elimisteve
// 2016.01.19

package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/tor"
	"github.com/elimisteve/cryptag/types"
	"github.com/elimisteve/fun"
	"github.com/stacktic/dropbox"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type DropboxRemote struct {
	dboxPath string
	rowsURL  string
	tagsURL  string

	dbox *dropbox.Dropbox

	cursorLock sync.RWMutex
	tagCursor  string // Used to fetch latest tags only

	// Used for encryption/decryption
	key *[32]byte

	dboxConf DropboxConfig
}

// SetTagCursor sets the cursor for the remote tags directory
// incremental TagPair fetching so it can be used for incremental
// TagPair fetching.
func (db *DropboxRemote) SetTagCursor(cursor string) {
	db.cursorLock.Lock()
	defer db.cursorLock.Unlock()

	db.tagCursor = cursor
}

// GetTagCursor gets the cursor for the remote tags directory used for
// incremental TagPair fetching.
func (db *DropboxRemote) GetTagCursor() string {
	db.cursorLock.RLock()
	defer db.cursorLock.RUnlock()

	return db.tagCursor
}

// SetHTTPClient sets the underlying HTTP client used. Probably most
// useful for using a custom client that does proxied requests,
// perhaps through Tor.
func (db *DropboxRemote) SetHTTPClient(c *http.Client) {
	// Goal: trigger
	// https://github.com/golang/oauth2/blob/master/internal/transport.go#L37-38
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, c)
	db.dbox.SetContext(ctx)
}

// UseTor sets db's HTTP client to one that uses Tor.
func (db *DropboxRemote) UseTor() error {
	c, err := tor.NewClient()
	if err != nil {
		return err
	}

	db.SetHTTPClient(c)

	if types.Debug {
		log.Println("*DropboxRemote to do HTTP calls over Tor")
	}

	return nil
}

func LoadDropboxRemote(backendPath, backendName string) (*DropboxRemote, error) {
	if backendPath == "" {
		backendPath = cryptag.BackendPath
	}
	if backendName == "" {
		host, _ := os.Hostname()
		backendName = "dropbox-" + host
	}
	backendName = strings.TrimSuffix(backendName, ".json")

	configFile := path.Join(backendPath, backendName+".json")

	if types.Debug {
		log.Printf("Reading backend config file `%v`\n", configFile)
	}

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

	dboxConf, err := DropboxConfigFromMap(conf.Custom)
	if err != nil {
		return nil, err
	}

	return NewDropboxRemote((*conf.Key)[:], conf.Name, dboxConf)
}

// NewDropboxRemote will save this backend to disk if len(key) == 0 or
// name == "".
func NewDropboxRemote(key []byte, name string, cfg DropboxConfig) (*DropboxRemote, error) {
	if cfg.BasePath != "/" {
		cfg.BasePath = strings.TrimRight(cfg.BasePath, "/")
	}

	if err := cfg.Valid(); err != nil {
		return nil, fmt.Errorf("Invalid token(s): %v", err)
	}

	dbox := dropbox.NewDropbox()
	dbox.SetAppInfo(cfg.AppKey, cfg.AppSecret)
	dbox.SetAccessToken(cfg.AccessToken)

	dboxPath := cfg.BasePath

	saveToDisk := false

	// Key

	if len(key) == 0 {
		saveToDisk = true
		wouldBeGoodKey, err := cryptag.RandomKey()
		if err != nil {
			return nil, err
		}
		// TODO(elimisteve): Really? Really?
		key = (*wouldBeGoodKey)[:]
	}
	goodKey, err := cryptag.ConvertKey(key)
	if err != nil {
		return nil, err
	}

	// Name

	if name == "" {
		host, _ := os.Hostname()
		name = "dropbox-" + host
		saveToDisk = true
	}

	if saveToDisk {
		config := &Config{
			Key:    goodKey,
			Name:   name,
			New:    true, // TODO(elimisteve): Make unnecessary; see filesystem.go
			Custom: DropboxConfigToMap(cfg),
		}
		if err = config.Canonicalize(); err != nil {
			return nil, err
		}
		if err = saveConfig(config); err != nil {
			return nil, err
		}
	}

	db := DropboxRemote{
		key:      goodKey,
		dboxPath: dboxPath,
		rowsURL:  path.Clean(dboxPath + "/rows"),
		tagsURL:  path.Clean(dboxPath + "/tags"),
		dbox:     dbox,
		dboxConf: cfg,
	}

	return &db, nil
}

func (db *DropboxRemote) ToConfig() (*Config, error) {
	if db.key == nil {
		return nil, cryptag.ErrNilKey
	}

	host, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("Error getting hostname for config Name: %v", err)
	}
	name := "dropbox-" + host

	config := Config{
		Key:    db.key,
		Name:   name,
		Custom: DropboxConfigToMap(db.dboxConf),
	}
	return &config, nil
}

func (db *DropboxRemote) Key() *[32]byte {
	return db.key
}

func (db *DropboxRemote) AllTagPairs() (types.TagPairs, error) {
	start := time.Now()

	pairs, err := getAllTagsFromDbox(db)
	if err != nil {
		return nil, err
	}
	if types.Debug {
		log.Printf("getAllTagsFromDbox took %v, returning %d TagPairs\n",
			time.Since(start), len(pairs))
	}

	return pairs, nil
}

func (db *DropboxRemote) SaveRow(row *types.Row) error {
	if len(row.Encrypted) == 0 || len(row.RandomTags) == 0 || row.Nonce == nil || *row.Nonce == [24]byte{} {
		if types.Debug {
			log.Printf("Error saving row `%#v`\n", row)
		}
		return errors.New("Invalid row; requires Encrypted, RandomTags, Nonce fields")
	}

	rowB, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("Error marshaling row: %v", err)
	}

	if types.Debug {
		log.Printf("POSTing row data: `%s`\n\n", rowB)
	}

	rclose := ioutil.NopCloser(bytes.NewReader(rowB))

	// TODO(elimisteve): Make sure that RandomTags aren't stored in
	// the file, too; storing them in the file _and_ filename would be
	// bad; should be filename only.
	dest := db.rowsURL + "/" + strings.Join(row.RandomTags, "-")

	_, err = db.dbox.FilesPut(rclose, int64(len(rowB)), dest, false, "")
	if err != nil {
		return err
	}

	return nil
}

func (db *DropboxRemote) SaveTagPair(pair *types.TagPair) error {
	pairB, err := json.Marshal(pair)
	if err != nil {
		return err
	}

	if types.Debug {
		log.Printf("POSTing tag pair data: `%s`\n\n", pairB)
	}

	rclose := ioutil.NopCloser(bytes.NewReader(pairB))
	dest := db.tagsURL + "/" + pair.Random

	_, err = db.dbox.FilesPut(rclose, int64(len(pairB)), dest, false, "")
	if err != nil {
		return err
	}

	if types.Debug {
		log.Printf("New *TagPair saved: `%#v`\n", pair)
	}

	return nil
}

func (db *DropboxRemote) TagPairsFromRandomTags(randtags cryptag.RandomTags) (types.TagPairs, error) {
	if len(randtags) == 0 {
		return nil, fmt.Errorf("Can't get 0 tags")
	}
	return getTagsFromDbox(db, randtags)
}

func (db *DropboxRemote) ListRows(randtags cryptag.RandomTags) (types.Rows, error) {
	includeFileBody := false
	return fetchRows(db, randtags, includeFileBody)
}

func (db *DropboxRemote) RowsFromRandomTags(randtags cryptag.RandomTags) (types.Rows, error) {
	includeFileBody := true
	return fetchRows(db, randtags, includeFileBody)
}

func (db *DropboxRemote) DeleteRows(randTags cryptag.RandomTags) error {
	return errors.New("DropboxRemote.DeleteRows NOT IMPLEMENTED")
}

//
// Helpers
//

func fetchRows(db *DropboxRemote, randtags cryptag.RandomTags, includeFileBody bool) (types.Rows, error) {
	query := strings.Join(randtags, " ")
	entries, err := db.dbox.Search(db.rowsURL, query, 0, false)
	if err != nil {
		return nil, err
	}

	return entriesToRows(db, entries, includeFileBody)
}

// getRowsFromDbox fetches the encrypted rows from url, decrypts them, then
func getRowsFromDbox(key *[32]byte, url string, pairs types.TagPairs) (types.Rows, error) {
	var rows types.Rows
	var err error

	if err = fun.FetchInto(url, HttpGetTimeout, &rows); err != nil {
		return nil, fmt.Errorf("Error from FetchInto: %v", err)
	}

	for _, row := range rows {
		if err = row.Populate(key, pairs); err != nil {
			return nil, fmt.Errorf("Error from PopulateRowAfterGet: %v", err)
		}
	}

	return rows, nil
}

func getAllTagsFromDbox(db *DropboxRemote) (types.TagPairs, error) {
	hash := db.GetTagCursor()
	if types.Debug {
		log.Printf("getAllTagsFromDbox: tag hash == `%v`\n", hash)
	}

	entry, err := db.dbox.Metadata(db.tagsURL, true, false, hash, "", 0)
	if err != nil {
		return nil, err
	}
	db.SetTagCursor(entry.Hash)

	randtags := make([]string, 0, len(entry.Contents))
	for i := range entry.Contents {
		randtags = append(randtags, filepath.Base(entry.Contents[i].Path))
	}

	return getTagsFromDbox(db, randtags)
}

// getTagsFromDbox fetches the encrypted tag pairs at db.tagsURL,
// decrypts them, and unmarshals them into a TagPairs value
func getTagsFromDbox(db *DropboxRemote, randtags cryptag.RandomTags) (types.TagPairs, error) {
	tags := make(chan *types.TagPair)

	// Download tags in randtags
	for _, tag := range randtags {
		go func(tag string) {
			pair, err := getTagFromDbox(db, tag)
			if err != nil {
				log.Printf("Error from getTagFromDbox: %v\n", err)
				tags <- nil
				return
			}
			tags <- pair
		}(tag)
	}

	var pairs types.TagPairs

	for i := 0; i < len(randtags); i++ {
		if t := <-tags; t != nil {
			// log.Printf("Tag #%d: %#v\n", i, t)
			pairs = append(pairs, t)
		}
	}

	if len(pairs) == 0 {
		log.Printf("getTagsFromDbox returning no pairs!\n")
	}

	return pairs, nil
}

func getTagFromDbox(db *DropboxRemote, tag string) (*types.TagPair, error) {
	b, err := download(db, db.tagsURL+"/"+tag)
	if err != nil {
		return nil, fmt.Errorf("Error from download: %v\n", err)
	}

	pair, err := newTagPair(b, tag)
	if err != nil {
		return nil, fmt.Errorf("Error from newTagPair: %v\n", err)
	}

	// Decrypt, thereby setting pair.plain
	if err = pair.Decrypt(db.Key()); err != nil {
		return nil, fmt.Errorf("Error from Decrypt: %v\n", err)
	}

	return pair, nil
}

func newTagPair(b []byte, filename string) (*types.TagPair, error) {
	var pair types.TagPair
	err := json.Unmarshal(b, &pair)
	if err != nil {
		return nil, err
	}
	pair.Random = filename

	return &pair, nil
}

func entriesToRows(db *DropboxRemote, entries []dropbox.Entry, includeFileBody bool) (types.Rows, error) {
	// Fetch rows
	rowCh := make(chan *types.Row)

	for _, entry := range entries {
		randtags := strings.Split(filepath.Base(entry.Path), "-")

		go func(entry dropbox.Entry) {
			// Row filename -> randtags

			var r *types.Row

			if includeFileBody {
				row, err := downloadRow(db, entry, randtags)
				if err != nil {
					log.Printf("Error from downloadRow: %v\n", err)
					rowCh <- nil
					return
				}
				r = row
			} else {
				r = &types.Row{RandomTags: randtags}
			}

			rowCh <- r
			return
		}(entry)
	}

	var rows types.Rows

	// Collect results (Rows successfully downloaded)
	for i := 0; i < len(entries); i++ {
		if r := <-rowCh; r != nil {
			rows = append(rows, r)
		}
	}

	return rows, nil
}

func stringsEqual(s, s2 []string) bool {
	if len(s) != len(s2) {
		return false
	}
	for i := range s {
		if s[i] != s2[i] {
			return false
		}
	}
	return true
}

func downloadRow(db *DropboxRemote, entry dropbox.Entry, randomTags []string) (*types.Row, error) {
	rowB, err := download(db, entry.Path)
	if err != nil {
		return nil, fmt.Errorf("Error downloading %v: %v\n", entry.Path, err)
	}

	row, err := types.NewRowFromBytes(rowB)
	if err != nil {
		return nil, fmt.Errorf("Error from NewRowFromBytes: %v\n", err)
	}

	if len(row.RandomTags) != 0 && !stringsEqual(randomTags, row.RandomTags) {
		log.Printf("PROBLEM: Row `%v` contains randtags `%v`!\n",
			entry.Path, row.RandomTags)
	}

	row.RandomTags = randomTags

	return row, nil
}

func download(db *DropboxRemote, fullURL string) (body []byte, err error) {
	if types.Debug {
		log.Printf("Downloading `%v`\n", fullURL)
	}
	f, _, err := db.dbox.Download(fullURL, "", 0)
	if err != nil {
		return nil, fmt.Errorf("Error downloading `%v`: %v\n", fullURL, err)
	}
	defer f.Close()

	// Read file
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Error from ReadAll: %v\n", err)
	}

	return b, nil
}
