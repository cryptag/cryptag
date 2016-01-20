// Steve Phillips / elimisteve
// 2015.08.08

package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/types"
	"github.com/elimisteve/fun"
	"github.com/elimisteve/help"
	"github.com/gorilla/mux"
)

var filesystem *FileSystem

func init() {
	fs, err := NewFileSystem(cryptag.Path)
	if err != nil {
		panic("Error from NewFileSystem: " + err.Error())
	}

	// Set global `filesystem` var
	filesystem = fs
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	router := mux.NewRouter()

	// Rows
	router.HandleFunc("/rows", GetRows).Methods("GET")
	router.HandleFunc("/rows", PostRow).Methods("POST")

	// Tags
	router.HandleFunc("/tags", GetTags).Methods("GET")
	router.HandleFunc("/tags", PostTag).Methods("POST")

	http.Handle("/", router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "7777"
	}
	server := fun.SimpleHTTPServer(router, ":"+port)

	log.Printf("HTTP server trying to listen on %v...\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func GetRows(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		help.WriteError(w, "Error parsing URL parameters: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	tags := req.Form["tags"]
	if len(tags) == 0 {
		log.Printf("No tags included; returning no rows")
		help.WriteError(w, "No tags included; returning no rows",
			http.StatusInternalServerError)
		return
	}

	// Tag format: /?tags=tag1,tag2,tag3
	tags = strings.Split(tags[0], ",")

	log.Printf("Rows queried by these tags: %+v\n", tags)
	rows, err := filesystem.RowsByTags(tags)
	if err != nil {
		help.WriteError(w, "Error fetching rows: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	log.Printf("%d Rows retrieved:\n%s", len(rows), rows)
	help.WriteJSON(w, rows)
}

func PostRow(w http.ResponseWriter, req *http.Request) {
	row := &types.Row{}
	if err := help.ReadInto(req.Body, row); err != nil {
		help.WriteError(w, "Error reading rows: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	err := filesystem.SaveRow(row)
	if err != nil {
		help.WriteError(w, "Error saving row: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	log.Printf("New row added: `%#v`\n", row)

	help.WriteJSON(w, row)
}

func GetTags(w http.ResponseWriter, req *http.Request) {
	// TODO(elimisteve): Should parse `?tags=t1,t2,t3` then use
	// filesytem.TagPairsFromRandomTags to find all TagPairs with
	// those random tags.  Currently we're ignoring these params and
	// just returning all TagPairs.
	allTagPairs, err := filesystem.AllTagPairs()
	if err != nil {
		help.WriteError(w, "Error getting TagPairs: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	help.WriteJSON(w, allTagPairs)
}

func PostTag(w http.ResponseWriter, req *http.Request) {
	pair := &types.TagPair{}

	err := help.ReadInto(req.Body, pair)
	if err != nil {
		http.Error(w, "Error reading tag pair: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	err = filesystem.SaveTagPair(pair)
	if err != nil {
		help.WriteError(w, "Error saving TagPair: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	log.Printf("New TagPair added: `%#v`\n", pair)

	help.WriteJSON(w, pair)
}

//
// TODO(elimisteve): Replace with pluggable server backends
//

type FileSystem struct {
	cryptagPath string
	tagsPath    string
	rowsPath    string
	dirLocks    map[string]sync.RWMutex

	randomTagToTagNonceSHA map[string]string
}

func NewFileSystem(cryptagPath string) (*FileSystem, error) {
	cryptagPath = strings.TrimRight(cryptagPath, "/\\")

	fs := &FileSystem{
		cryptagPath: cryptagPath,
		tagsPath:    path.Join(cryptagPath, "tags"),
		rowsPath:    path.Join(cryptagPath, "rows"),
		dirLocks:    map[string]sync.RWMutex{},
	}
	if err := fs.Init(); err != nil {
		return nil, err
	}
	if err := fs.setRandomTagToTagNonceSHA(); err != nil {
		return nil, err
	}

	return fs, nil
}

// Init creates the base CrypTag directories
func (fs *FileSystem) Init() error {
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

func (fs *FileSystem) SaveTagPair(pair *types.TagPair) error {
	if len(pair.PlainEncrypted) == 0 || len(pair.Random) == 0 || pair.Nonce == nil || *pair.Nonce == [24]byte{} {
		// TODO(elimisteve): Make error global?
		return errors.New("Invalid tag pair; requires plain_encrypted, random, and nonce fields")
	}

	// Save tag pair to fs.tagsPath/$nonce_sha/{plain_encrypted,random,nonce}
	tagNonceSHA := nonceSHA256(pair.Nonce)
	dirPath := path.Join(fs.tagsPath, tagNonceSHA)
	if err := os.Mkdir(dirPath, 0755); err != nil && !os.IsExist(err) {
		// TODO(elimisteve): Sanitize?
		return fmt.Errorf("Error making dir `%s`: %v", dirPath, err)
	}

	// Make files in dirPath

	// nonce
	nonceF, err := os.Create(path.Join(dirPath, "nonce"))
	if err != nil {
		return fmt.Errorf("Error creating file `nonce`: %v", err)
	}
	defer nonceF.Close()
	if _, err = nonceF.Write((*pair.Nonce)[:]); err != nil {
		return fmt.Errorf("Error saving nonce: %v", err)
	}

	// random
	randomF, err := os.Create(path.Join(dirPath, "random"))
	if err != nil {
		return fmt.Errorf("Error creating file `random`: %v", err)
	}
	defer randomF.Close()
	if _, err = randomF.Write([]byte(pair.Random)); err != nil {
		return fmt.Errorf("Error saving random string: %v", err)
	}

	// plain_encrypted
	plainEncF, err := os.Create(path.Join(dirPath, "plain_encrypted"))
	if err != nil {
		return fmt.Errorf("Error creating file `plain_encrypted`: %v", err)
	}
	defer plainEncF.Close()
	if _, err = plainEncF.Write(pair.PlainEncrypted); err != nil {
		return fmt.Errorf("Error saving plain_encrypted: %v", err)
	}

	// Saved!
	return nil
}

func (fs *FileSystem) AllTagPairs() (types.TagPairs, error) {
	nonceDirs, err := filepath.Glob(path.Join(fs.tagsPath, "*"))
	if err != nil {
		return nil, fmt.Errorf("Error listing tags: %v", err)
	}

	var pairs types.TagPairs
	for _, dir := range nonceDirs {
		// Read plain_encrypted, random, and nonce
		pair, err := readTagFiles(dir)
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, pair)
	}

	return pairs, nil
}

func readTagFiles(tagDir string) (*types.TagPair, error) {
	// TODO(elimisteve): Do streaming reads
	plainEncF, err := os.Open(path.Join(tagDir, "plain_encrypted"))
	if err != nil {
		return nil, err
	}
	defer plainEncF.Close()

	randomF, err := os.Open(path.Join(tagDir, "random"))
	if err != nil {
		return nil, err
	}
	defer randomF.Close()

	nonceF, err := os.Open(path.Join(tagDir, "nonce"))
	if err != nil {
		return nil, err
	}
	defer nonceF.Close()

	// Read files and create new TagPair out of them

	// plain_encrypted
	plainEnc, err := ioutil.ReadAll(plainEncF)
	if err != nil {
		return nil, err
	}

	// random
	randomBody, err := ioutil.ReadAll(randomF)
	if err != nil {
		return nil, err
	}
	random := string(randomBody)

	// nonce
	nonceBody, err := ioutil.ReadAll(nonceF)
	if err != nil {
		return nil, err
	}
	nonce, err := cryptag.ConvertNonce(nonceBody)
	if err != nil {
		return nil, err
	}

	return types.NewTagPair(plainEnc, random, nonce, ""), nil
}

func (fs *FileSystem) SaveRow(row *types.Row) error {
	if len(row.Encrypted) == 0 || len(row.RandomTags) == 0 || row.Nonce == nil || *row.Nonce == [24]byte{} {
		// TODO(elimisteve): Make error global?
		return errors.New("Invalid row; requires data, tags, and nonce fields")
	}

	// Save tag row to fs.tagsPath/$nonce_sha/{plain_encrypted,random,nonce}
	rowNonceSHA := nonceSHA256(row.Nonce)
	dirPath := path.Join(fs.rowsPath, rowNonceSHA)
	if err := os.Mkdir(dirPath, 0755); err != nil && !os.IsExist(err) {
		// TODO(elimisteve): Sanitize?
		return fmt.Errorf("Error making dir `%s`: %v", dirPath, err)
	}

	// Make files in dirPath

	// nonce
	nonceF, err := os.Create(path.Join(dirPath, "nonce"))
	if err != nil {
		return fmt.Errorf("Error creating file `nonce`: %v", err)
	}
	defer nonceF.Close()
	if _, err = nonceF.Write((*row.Nonce)[:]); err != nil {
		return fmt.Errorf("Error saving nonce: %v", err)
	}

	// data
	encF, err := os.Create(path.Join(dirPath, "data"))
	if err != nil {
		return fmt.Errorf("Error creating file `data`: %v", err)
	}
	defer encF.Close()
	if _, err = encF.Write(row.Encrypted); err != nil {
		return fmt.Errorf("Error saving data: %v", err)
	}

	// Create tags/ dir and symlink from
	// fs.tagsPath/$tag_nonce_sha to
	// fs.rowsPath/$row_nonce_sha/tags/$random_tag
	rowTagsPath := path.Join(fs.rowsPath, rowNonceSHA, "tags")
	if err := os.Mkdir(rowTagsPath, 0755); err != nil && !os.IsExist(err) {
		// TODO(elimisteve): Sanitize?
		return fmt.Errorf("Error making dir `%s`: %v", dirPath, err)
	}

	err = fs.setRandomTagToTagNonceSHA()
	if err != nil {
		return fmt.Errorf("Error looking up latest tags: %v", err)
	}

	// Create symlinks to tag rows
	for _, randomTag := range row.RandomTags {
		tagNonceSHA, ok := fs.randomTagToTagNonceSHA[randomTag]
		if !ok {
			return fmt.Errorf("randomTag `%s` not found in `%v` of length %d",
				randomTag, fs.randomTagToTagNonceSHA, len(fs.randomTagToTagNonceSHA))
		}

		// From {~/.cryptag/tags}/$tag_nonce_sha
		from := path.Join(fs.tagsPath, tagNonceSHA)

		// To {~/.cryptag/rows/$row_nonce_sha/tags}/$random_tag
		to := path.Join(rowTagsPath, randomTag)

		if err = os.Symlink(from, to); err != nil {
			return err
		}
	}

	// Saved!
	return nil
}

func (fs *FileSystem) TagPairsFromRandomTags(randtags []string) (types.TagPairs, error) {
	return nil, fmt.Errorf("Error TagPairsFromRandomTags NOT IMPLEMENTED")
}

func (fs *FileSystem) RowsByTags(randomTags []string) (types.Rows, error) {
	// Find the rows that have all the tags listed in randomTags

	rowDirs, err := filepath.Glob(path.Join(fs.rowsPath, "*"))
	if err != nil {
		return nil, err
	}

	if len(randomTags) == 0 {
		var rows types.Rows
		for _, dir := range rowDirs {
			row, err := readRowFiles(dir)
			if err != nil {
				return nil, err
			}

			rows = append(rows, row)
		}

		return rows, nil
	}

	// len(randomTag) > 0

	var rows types.Rows

	// For each row dir, if it has all tags, save it
	for _, dir := range rowDirs {
		rowTags, err := filepath.Glob(path.Join(dir, "tags", "*"))
		if err != nil {
			return nil, err
		}

		// We just need the filenames, not the full paths
		rowTags = filenames(rowTags)

		if !fun.SliceContainsAll(rowTags, randomTags) {
			log.Printf("`%v`  doesn't contain all tags in  `%v`\n", rowTags, randomTags)
			continue
		}

		// Row is tagged with all randomTags; return to user

		row, err := readRowFiles(dir)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row)
	}

	return rows, nil
}

//
// Helpers
//

func readRowFiles(rowDir string) (*types.Row, error) {
	// Open, read ./data
	data, err := ioutil.ReadFile(path.Join(rowDir, "data"))
	if err != nil {
		return nil, err
	}
	row := types.Row{Encrypted: data}

	// Open, read ./nonce
	nonceB, err := ioutil.ReadFile(path.Join(rowDir, "nonce"))
	if err != nil {
		return nil, err
	}
	nonce, err := cryptag.ConvertNonce(nonceB)
	if err != nil {
		return nil, err
	}

	row.Nonce = nonce

	// List files in ./tags to get random tags
	randomTags, err := filepath.Glob(path.Join(rowDir, "tags", "*"))
	if err != nil {
		return nil, err
	}

	// Attach random tags to row
	row.RandomTags = filenames(randomTags)

	return &row, nil
}

func filenames(paths []string) (filenames []string) {
	filenames = make([]string, 0, len(paths))
	for _, path := range paths {
		filenames = append(filenames, filepath.Base(path))
	}
	return filenames
}

func (fs *FileSystem) setRandomTagToTagNonceSHA() error {
	// Search all tagsPath/$tag_nonce_sha/random files for body == randomTag
	randoms, err := filepath.Glob(path.Join(fs.tagsPath, "*", "random"))
	if err != nil {
		return err
	}

	// map[randomTag]TagNonceSHA
	m := make(map[string]string, len(randoms))

	for _, rfilename := range randoms {
		f, err := os.Open(rfilename)
		if err != nil {
			return err
		}

		randomTag, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		// rfilename == fs.tagsPath/$tag_nonce_sha/random

		// Trim `$fs.tagsPath/` off the front
		rfilename = rfilename[len(fs.tagsPath)+1:]

		// Trim `/random` off the end
		tagNonceSHA := rfilename[:len(rfilename)-len("random")-1]

		m[string(randomTag)] = tagNonceSHA
	}

	fs.randomTagToTagNonceSHA = m

	return nil
}

func nonceSHA256(nonce *[24]byte) string {
	return fmt.Sprintf("%x", sha256.Sum256((*nonce)[:]))
}
