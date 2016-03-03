// Steve Phillips / elimisteve
// 2015.08.08

package main

import (
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

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/types"
	"github.com/elimisteve/fun"
	"github.com/elimisteve/help"
	"github.com/gorilla/mux"
)

var filesystem *FileSystem

func init() {
	fs, err := NewFileSystem(cryptag.LocalDataPath)
	if err != nil {
		log.Fatalf("Error from NewFileSystem: %v", err)
	}

	// Set global `filesystem` var
	filesystem = fs
}

func main() {
	router := mux.NewRouter()

	// Rows
	router.HandleFunc("/", GetRoot).Methods("GET")
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

func GetRoot(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(`Welcome to CrypTag!

You can use CrypTag Sandstorm grains for storing passwords or other
secrets. The three best things about CrypTag:

- You access all your data from a secure client on your own computer.

- Data never goes out of sync. (All data is stored on the server.)

- Searches are efficient. This is CrypTag's key idea: store secret information on a server,
  with labels ("tags") that the server can't understand, that can still be used for search!


Overview of daily use
---------------------

$ cpass-sandstorm create mytw1tt3rp4ssword twitter @myusername login:myusername

This stores your Twitter password (in encrypted form, of course) to Sandstorm.


$ cpass-sandstorm @myusername

This adds the Twitter password for @myusername to your clipboard automatically!


$ cpass-sandstorm all

This will list all passwords (and, actually, all other textual data;
see below) you've stored.


More tips/use cases
-------------------

cpass-sandstorm is for more than just passwords, though.  You may also
want to store and access:

1. Credit card numbers (cpass-sandstorm visa digits)
2. Quotes (cpass-sandstorm nietzsche quote)
3. Bookmarks, tagged like on Pinboard or Delicious (cpass-sandstorm url snowden)
4. Command line commands -- cross-machine shell history! (cpass-sandstorm install docker)
5. GitHub recovery codes (cpass-sandstorm github recoverycode)


Get started
-----------
	
Download and run the cpass-sandstorm Linux command line client:

$ mkdir ~/bin; cd ~/bin && wget https://github.com/elimisteve/cryptag/blob/master/bin/cpass-sandstorm?raw=true -O cpass-sandstorm && chmod +x cpass-sandstorm && ./cpass-sandstorm

Then click the Key icon above this message (on Sandstorm) and generate a Sandstorm API key to give to cpass-sandstorm like so:

$ ./cpass-sandstorm init <sandstorm_key>

To see the remaining valid commands (such as "create", seen above), run

$ ./cpass-sandstorm

Enjoy!


Learn more
----------

You'll find more details at:

- Conceptual overview in these slides from my DEFCON talk introducing CrypTag: https://talks.stevendphillips.com/cryptag-defcon23-cryptovillage/

- Details: https://github.com/elimisteve/cryptag
`))
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
			http.StatusBadRequest)
		return
	}

	// Tag format: /?tags=tag1,tag2,tag3
	tags = strings.Split(tags[0], ",")

	if types.Debug {
		log.Printf("Rows queried by these tags: %+v\n", tags)
	}

	includeFileBody := true
	rows, err := filesystem.RowsByTags(tags, includeFileBody)
	if err != nil {
		help.WriteError(w, "Error fetching rows: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

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

	if types.Debug {
		log.Printf("New row added: `%#v`\n", row)
	}

	help.WriteJSON(w, row)
}

func GetTags(w http.ResponseWriter, req *http.Request) {
	_ = req.ParseForm()

	// TODO(elimisteve): Should use filesytem.TagPairsFromRandomTags
	// to find all TagPairs with those random tags.
	allTagPairs, err := filesystem.AllTagPairs()
	if err != nil {
		help.WriteError(w, "Error getting TagPairs: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	randtags := req.Form["tags"]
	if len(randtags) == 0 {
		if types.Debug {
			log.Printf("All %d TagPairs retrieved", len(allTagPairs))
		}
		help.WriteJSON(w, allTagPairs)
		return
	}
	randtags = strings.Split(randtags[0], ",")

	pairs, _ := allTagPairs.HaveAllRandomTags(randtags)
	help.WriteJSON(w, pairs)
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

	if types.Debug {
		log.Printf("New TagPair added: `%#v`\n", pair)
	}

	help.WriteJSON(w, pair)
}

//
// TODO(elimisteve): Replace with pluggable server backends
//

type FileSystem struct {
	cryptagPath string
	tagsPath    string
	rowsPath    string
}

func NewFileSystem(cryptagPath string) (*FileSystem, error) {
	cryptagPath = strings.TrimRight(cryptagPath, "/\\")

	fs := &FileSystem{
		cryptagPath: cryptagPath,
		tagsPath:    path.Join(cryptagPath, "tags"),
		rowsPath:    path.Join(cryptagPath, "rows"),
	}
	if err := fs.Init(); err != nil {
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

	filename := path.Join(fs.tagsPath, pair.Random)

	return ioutil.WriteFile(filename, b, 0644)
}

func (fs *FileSystem) AllTagPairs() (types.TagPairs, error) {
	tagFiles, err := filepath.Glob(path.Join(fs.tagsPath, "*"))
	if err != nil {
		return nil, fmt.Errorf("Error listing tags: %v", err)
	}

	var pairs types.TagPairs
	for _, f := range tagFiles {
		// Read plain_encrypted, random, and nonce
		pair, err := readTagFile(f)
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, pair)
	}

	return pairs, nil
}

func readTagFile(tagFile string) (*types.TagPair, error) {
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

	return pair, nil
}

func (fs *FileSystem) SaveRow(row *types.Row) error {
	if len(row.Encrypted) == 0 || len(row.RandomTags) == 0 || row.Nonce == nil || *row.Nonce == [24]byte{} {
		// TODO(elimisteve): Make error global?
		return errors.New("Invalid row; requires data, tags, and nonce fields")
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

	// Create row file fs.rowsPath/randomtag1-randomtag2-randomtag3

	filename := path.Join(fs.rowsPath, strings.Join(row.RandomTags, "-"))

	return ioutil.WriteFile(filename, b, 0644)
}

func (fs *FileSystem) TagPairsFromRandomTags(randtags []string) (types.TagPairs, error) {
	return nil, fmt.Errorf("Error TagPairsFromRandomTags NOT IMPLEMENTED")
}

func (fs *FileSystem) RowsByTags(randTags []string, includeFileBody bool) (types.Rows, error) {
	if types.Debug {
		log.Printf("RowsByTags(%#v, %v)\n", randTags, includeFileBody)
	}

	rowFiles, err := filepath.Glob(path.Join(fs.rowsPath, "*"))
	if err != nil {
		return nil, err
	}

	var rows types.Rows

	// For each row dir, if it has all tags, append to `rows`
	for _, rowFile := range rowFiles {
		// Row filenames are of the form randtag1-randtag2-randtag3
		rowTags := strings.Split(filepath.Base(rowFile), "-")

		if !fun.SliceContainsAll(rowTags, randTags) {
			continue
		}

		// Row is tagged with all queryTags; return to user
		row := &types.Row{RandomTags: rowTags}

		// Load contents of row file, too
		if includeFileBody {
			row, err = readRowFile(rowFile, rowTags)
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

//
// Helpers
//

func readRowFile(rowFilePath string, rowTags []string) (*types.Row, error) {
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
