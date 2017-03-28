// Steve Phillips / elimisteve
// 2015.11.04

package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/types"
	"github.com/elimisteve/fun"
)

var (
	ErrWrongBackendType = errors.New("backend: wrong Backend type")
)

type FileSystem struct {
	name     string
	dataPath string
	tagsPath string // subdirectory of dataPath
	rowsPath string // subdirectory of dataPath
	new      bool
	key      *[32]byte
}

func NewFileSystem(conf *Config) (*FileSystem, error) {
	if err := conf.Canonicalize(); err != nil {
		return nil, err
	}

	fs := &FileSystem{
		name:     conf.Name,
		dataPath: conf.DataPath,
		tagsPath: path.Join(conf.DataPath, "tags"),
		rowsPath: path.Join(conf.DataPath, "rows"),
		new:      conf.New,
		key:      conf.Key,
	}
	if err := fs.init(); err != nil {
		return nil, err
	}

	// Save config to disk
	if conf.New {
		if err := saveConfig(conf); err != nil {
			return nil, err
		}
	}

	return fs, nil
}

func saveConfig(conf *Config) error {
	return conf.Save(cryptag.BackendPath)
}

// init creates the base CrypTag directories
func (fs *FileSystem) init() error {
	var err error
	// TODO(elimisteve): Should this assume that cryptag.BackendPath
	// already exists?
	for _, path := range []string{fs.dataPath, fs.tagsPath, fs.rowsPath, cryptag.BackendPath} {
		err = os.MkdirAll(path, 0755)
		if err == nil || os.IsExist(err) {
			// Created successfully or already exists
			continue
		}
		return fmt.Errorf("Error making dir `%s`: %v", path, err)
	}
	return nil
}

func LoadOrCreateFileSystem(backendPath, backendName string) (*FileSystem, error) {
	if backendName == "" {
		backendName, _ = os.Hostname()
	}

	conf, err := ReadConfig(backendPath, backendName)
	if err != nil {
		// If config doesn't exist, create new one

		if os.IsNotExist(err) {
			conf = &Config{
				Name:  backendName,
				Type:  TypeFileSystem,
				New:   true,
				Local: true,
			}
			return NewFileSystem(conf)
		}
		return nil, err
	}

	if conf.GetType() != TypeFileSystem {
		return nil, ErrWrongBackendType
	}

	return NewFileSystem(conf)
}

// LoadOrCreateDefaultFileSystemBackend calls LoadOrCreateFileSystem
// then, if the returned FileSystem was just created anew, makes it
// the default Backend.
func LoadOrCreateDefaultFileSystemBackend(backendPath, backendName string) (*FileSystem, error) {
	fs, err := LoadOrCreateFileSystem(backendPath, backendName)
	if err != nil {
		return nil, err
	}
	if fs.new {
		err = SetDefaultBackend(backendPath, fs.name)
		if err != nil {
			return fs, err
		}
	}
	return fs, nil
}

func (fs *FileSystem) Name() string {
	return fs.name
}

func (fs *FileSystem) ToConfig() (*Config, error) {
	name := fs.name

	if name == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("Error getting hostname: %v", err)
		}
		name = hostname
	}

	config := Config{
		Name:     name,
		Type:     TypeFileSystem,
		New:      fs.new,
		Key:      fs.key,
		DataPath: fs.dataPath,
	}

	return &config, nil
}

func (fs *FileSystem) Key() *[32]byte {
	return fs.key
}

func (fs *FileSystem) AllTagPairs(oldPairs types.TagPairs) (types.TagPairs, error) {
	tagFiles, err := filepath.Glob(path.Join(fs.tagsPath, "*"))
	if err != nil {
		return nil, fmt.Errorf("Error listing tags: %v", err)
	}

	var pairs types.TagPairs
	for _, f := range tagFiles {
		// filepath.Base(f) is of the form randtag1-randtag2-randtag3
		// and its contents is {"plain_encrypted": ..., "nonce": ...}
		pair, err := readTagFile(fs.Key(), f)
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, pair)
	}

	if types.Debug {
		log.Printf("AllTagPairs: returning %d pairs (%d just fetched)\n",
			len(pairs), len(pairs)-len(oldPairs))
	}

	return pairs, nil
}

func (fs *FileSystem) TagPairsFromRandomTags(randtags cryptag.RandomTags) (types.TagPairs, error) {
	// TODO: This should grab files from disk with the names $BASE/tags/$randtag
	return nil, errors.New("TagPairsFromRandomTags: NOT IMPLEMENTED")
}

func (fs *FileSystem) SaveTagPair(pair *types.TagPair) error {
	if len(pair.PlainEncrypted) == 0 || len(pair.Random) == 0 || pair.Nonce == nil || *pair.Nonce == [24]byte{} {
		// TODO(elimisteve): Make error global?
		return errors.New("Invalid tag pair; requires plain_encrypted, random, and nonce fields")
	}

	// Just save "plain_encrypted" and "nonce" to file ("random"
	// contained in filename)
	t := map[string]interface{}{
		"plain_encrypted": pair.PlainEncrypted,
		"nonce":           pair.Nonce,
	}
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}

	// Save tag pair to fs.tagsPath/$random
	filepath := path.Join(fs.tagsPath, pair.Random)

	return ioutil.WriteFile(filepath, b, 0600)
}

func (fs *FileSystem) ListRows(randtags cryptag.RandomTags) (types.Rows, error) {
	// TODO: Reduce code duplication between ListRows and
	// RowsFromPlainTags

	if len(randtags) == 0 {
		return nil, errors.New("Must query by 1 or more tags")
	}

	// len(randtags) > 0

	return fs.rowsFromRandomTags(randtags, false)
}

func (fs *FileSystem) RowsFromRandomTags(randtags cryptag.RandomTags) (types.Rows, error) {
	if len(randtags) == 0 {
		return nil, errors.New("Must query by 1 or more tags")
	}

	// Find the rows that have all the tags in plainTags

	// len(randtags) > 0

	return fs.rowsFromRandomTags(randtags, true)
}

func (fs *FileSystem) SaveRow(row *types.Row) error {
	if len(row.Encrypted) == 0 || len(row.RandomTags) == 0 || row.Nonce == nil || *row.Nonce == [24]byte{} {
		if types.Debug {
			log.Printf("Error saving row `%#v`\n", row)
		}
		// TODO(elimisteve): Make error global?
		return errors.New("Invalid row; requires Encrypted, RandomTags, Nonce fields")
	}

	// Save row.{Encrypted,Nonce} to fs.rowsPath/randomtag1-randomtag2-randomtag3

	rowData := map[string]interface{}{
		"data":  row.Encrypted,
		"nonce": row.Nonce,
	}
	b, err := json.Marshal(rowData)
	if err != nil {
		return err
	}

	// Create row file fs.rowsPath/randomtag1-randomtag2-randomtag3-...

	filename := strings.Join(row.RandomTags, "-")
	filepath := path.Join(fs.rowsPath, filename)

	return ioutil.WriteFile(filepath, b, 0600)
}

func (fs *FileSystem) DeleteRows(randTags cryptag.RandomTags) error {
	if len(randTags) == 0 {
		return fmt.Errorf("Must query by 1 or more tags")
	}

	if types.Debug {
		log.Printf("DeleteRows(%#v)\n", randTags)
	}

	// Find rows matching given tags
	rows, err := fs.rowsFromRandomTags(randTags, false)
	if err != nil {
		return err
	}

	if types.Debug {
		log.Printf("DeleteRows: deleting %d rows: %s\n", len(rows), rows)
	}

	// Delete
	for _, row := range rows {
		filename := path.Join(fs.rowsPath, strings.Join(row.RandomTags, "-"))
		if types.Debug {
			log.Printf("Removing row file `%v`\n", filename)
		}
		err = os.Remove(filename)
		if err != nil {
			return err
		}
	}

	return nil
}

//
// Helpers
//

func (fs *FileSystem) rowsFromRandomTags(randTags []string, includeFileBody bool) (types.Rows, error) {
	if types.Debug {
		log.Printf("rowsFromRandomTags(%#v, %v)\n", randTags, includeFileBody)
	}

	rowFiles, err := filepath.Glob(path.Join(fs.rowsPath, "*"))
	if err != nil {
		return nil, err
	}

	var rows types.Rows

	// For each row dir, if it has all tags, append to `rows`
	for _, f := range rowFiles {
		// Row filenames are of the form randtag1-randtag2-randtag3
		rowTags := strings.Split(filepath.Base(f), "-")

		if !fun.SliceContainsAll(rowTags, randTags) {
			continue
		}

		var row *types.Row

		// Load contents of row file, too
		if includeFileBody {
			row, err = readRowFile(fs, f, rowTags)
			if err != nil {
				return nil, err
			}
		} else {
			// Row is tagged with all queryTags; return to user
			row = &types.Row{RandomTags: rowTags}
		}

		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil, types.ErrRowsNotFound
	}

	return rows, nil
}

func readTagFile(key *[32]byte, tagFile string) (*types.TagPair, error) {
	// TODO(elimisteve): Do streaming reads

	// Set pair.{PlainEncrypted,Nonce} from file contents, pair.Random
	// from filename

	// File contains: {"plain_encrypted": ..., "nonce": ...}
	b, err := ioutil.ReadFile(tagFile)
	if err != nil {
		return nil, err
	}

	pair := &types.TagPair{}
	err = json.Unmarshal(b, pair)
	if err != nil {
		return nil, err
	}

	pair.Random = filepath.Base(tagFile)

	// Populate pair.plain
	if err = pair.Decrypt(key); err != nil {
		return nil, fmt.Errorf("Error from pair.Decrypt: %v", err)
	}

	return pair, nil
}

func readRowFile(bk *FileSystem, rowFilePath string, rowTags []string) (*types.Row, error) {
	b, err := ioutil.ReadFile(rowFilePath)
	if err != nil {
		return nil, err
	}

	var row types.Row
	// This populates row.Encrypted and row.Nonce
	err = json.Unmarshal(b, &row)
	if err != nil {
		return nil, err
	}

	row.RandomTags = rowTags

	return &row, nil
}
