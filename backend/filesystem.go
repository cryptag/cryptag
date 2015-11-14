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

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/types"
	"github.com/elimisteve/fun"
)

type FileSystem struct {
	cryptagPath string
	tagsPath    string
	rowsPath    string
	key         *[32]byte
}

func NewFileSystem(conf *Config) (*FileSystem, error) {
	if err := conf.Canonicalize(); err != nil {
		return nil, err
	}

	fs := &FileSystem{
		cryptagPath: conf.CryptagBasePath,
		tagsPath:    path.Join(conf.CryptagBasePath, "tags"),
		rowsPath:    path.Join(conf.CryptagBasePath, "rows"),
		key:         conf.Key,
	}
	if err := fs.init(); err != nil {
		return nil, err
	}

	return fs, nil
}

// init creates the base CrypTag directories
func (fs *FileSystem) init() error {
	var err error
	for _, path := range []string{fs.cryptagPath, fs.tagsPath, fs.rowsPath} {
		err = os.MkdirAll(path, 0755)
		if err == nil || os.IsExist(err) {
			// Created successfully or already exists
			continue
		}
		return fmt.Errorf("Error making dir `%s`: %v", path, err)
	}
	return nil
}

func (fs *FileSystem) Encrypt(plain []byte, nonce *[24]byte) ([]byte, error) {
	return cryptag.Encrypt(plain, nonce, fs.key)
}

func (fs *FileSystem) Decrypt(cipher []byte, nonce *[24]byte) (plain []byte, err error) {
	return cryptag.Decrypt(cipher, nonce, fs.key)
}

func (fs *FileSystem) AllTagPairs() (types.TagPairs, error) {
	tagFiles, err := filepath.Glob(path.Join(fs.tagsPath, "*"))
	if err != nil {
		return nil, fmt.Errorf("Error listing tags: %v", err)
	}

	var pairs types.TagPairs
	for _, f := range tagFiles {
		// filepath.Base(f) is of the form randtag1-randtag2-randtag3
		// and its contents is {"plain_encrypted": ..., "nonce": ...}
		pair, err := readTagFile(fs, f)
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, pair)
	}

	return pairs, nil
}

func (fs *FileSystem) TagPairsFromRandomTags(randTags []string) (types.TagPairs, error) {
	pairs, err := fs.AllTagPairs()
	if err != nil {
		return nil, err
	}
	return pairs.HaveAllRandomTags(randTags)
}

func (fs *FileSystem) SaveTagPair(pair *types.TagPair) (*types.TagPair, error) {
	if len(pair.PlainEncrypted) == 0 || len(pair.Random) == 0 || pair.Nonce == nil || *pair.Nonce == [24]byte{} {
		// TODO(elimisteve): Make error global?
		return nil, errors.New("Invalid tag pair; requires plain_encrypted, random, and nonce fields")
	}

	// Just save "plain_encrypted" and "nonce" to file ("random"
	// contained in filename)
	t := map[string]interface{}{
		"plain_encrypted": pair.PlainEncrypted,
		"nonce":           pair.Nonce,
	}
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	// Save tag pair to fs.tagsPath/$random
	tagpairF, err := os.Create(path.Join(fs.tagsPath, pair.Random))
	if err != nil {
		return nil, fmt.Errorf("Error creating tag file: %v", err)
	}
	defer tagpairF.Close()

	if _, err = tagpairF.Write(b); err != nil {
		return nil, fmt.Errorf("Error saving tag file: %v", err)
	}

	// Saved!
	return pair, nil
}

func (fs *FileSystem) RowsFromPlainTags(plainTags []string) (types.Rows, error) {
	if len(plainTags) == 0 {
		return nil, errors.New("Must query by 1 or more tags")
	}

	// Find the rows that have all the tags in plainTags

	queryTags, err := randomFromPlain(fs, plainTags)
	if err != nil {
		return nil, err
	}

	// len(queryTags) > 0

	rows, err := fs.rowsFromRandomTags(queryTags, true)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (fs *FileSystem) SaveRow(r *types.Row) (*types.Row, error) {
	row, err := PopulateRowBeforeSave(fs, r)
	if err != nil {
		return nil, err
	}

	if len(row.Encrypted) == 0 || len(row.RandomTags) == 0 || row.Nonce == nil || *row.Nonce == [24]byte{} {
		if types.Debug {
			log.Printf("Error saving row `%#v`\n", row)
		}
		// TODO(elimisteve): Make error global?
		return nil, errors.New("Invalid row; requires data, tags, and nonce fields")
	}

	// Save row.{Encrypted,Nonce} to fs.rowsPath/randomtag1-randomtag2-randomtag3

	rowData := map[string]interface{}{
		"data":  row.Encrypted,
		"nonce": row.Nonce,
	}
	b, err := json.Marshal(rowData)
	if err != nil {
		return nil, err
	}

	// Create row file fs.rowsPath/randomtag1-randomtag2-randomtag3
	rowF, err := os.Create(path.Join(fs.rowsPath, strings.Join(row.RandomTags, "-")))
	if err != nil {
		return nil, fmt.Errorf("Error creating row file: %v", err)
	}
	defer rowF.Close()
	if _, err = rowF.Write(b); err != nil {
		return nil, fmt.Errorf("Error saving row file: %v", err)
	}

	// Saved!
	return row, nil
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

		// Row is tagged with all queryTags; return to user
		row := &types.Row{RandomTags: rowTags}

		// Load contents of row file, too
		if includeFileBody {
			row, err = readRowFile(fs, f, rowTags)
			if err != nil {
				return nil, err
			}
		}

		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil, types.ErrRowsNotFound
	}

	return rows, nil
}

func readTagFile(fs *FileSystem, tagFile string) (*types.TagPair, error) {
	// TODO(elimisteve): Do streaming reads

	// Set pair.{PlainEncrypted,Nonce} from file contents, pair.Random
	// from filename

	// File contains: {"plain_encrypted": ..., "nonce": ...}
	b, err := openAndRead(tagFile)
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
	if err = pair.Decrypt(fs.Decrypt); err != nil {
		return nil, fmt.Errorf("Error from pair.Decrypt: %v", err)
	}

	return pair, nil
}

func readRowFile(fs *FileSystem, rowFilePath string, rowTags []string) (*types.Row, error) {
	b, err := openAndRead(rowFilePath)
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

	// Populate row.decrypted and row.plain
	if err = PopulateRowAfterGet(fs, &row); err != nil {
		return nil, err
	}

	return &row, nil
}

func openAndRead(filename string) (contents []byte, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}
