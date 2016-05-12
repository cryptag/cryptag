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

const (
	deleteRowDelete = "delete"
	deleteRowMove   = "move"
)

var deleteOptions = []string{
	deleteRowDelete,
	deleteRowMove,
}

var onRowDelete string // defines what happens when Row is deleted

func init() {
	onDelete := os.Getenv("ON_DELETE")
	if onDelete == "" {
		onDelete = deleteRowMove
	}

	if !fun.SliceContains(deleteOptions, onDelete) {
		log.Fatalf("ON_DELETE env var set to invalid option `%s`\n", onDelete)
	}

	// Set global `onRowDelete` var
	onRowDelete = onDelete

	log.Printf("Row deletion behavior: %s\n", onRowDelete)
}

func main() {
	router := mux.NewRouter()

	// Rows
	router.HandleFunc("/", GetRoot).Methods("GET")
	router.HandleFunc("/rows", GetRows).Methods("GET")
	router.HandleFunc("/rows", PostRow).Methods("POST")
	router.HandleFunc("/rows/list", ListRows).Methods("GET")
	router.HandleFunc("/rows/delete", DeleteRows).Methods("GET")

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

You can use CrypTag Sandstorm grains for storing passwords, files, or
other data. The three best things about CrypTag are:

1. You access all your data from a secure client on your own computer.

2. Data never goes out of sync. (All data is stored on the server.)

3. Searches are efficient. This is CrypTag's key idea: store secret
information on a server, with tags that the server can't understand,
that can still be used for search!


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


Getting started
---------------

## Linux and Mac OS X

Run this to download the cpass-sandstorm command line program:

    $ mkdir ~/bin; cd ~/bin && C="cpass-sandstorm" && curl -SL https://github.com/elimisteve/cryptag/blob/v1-beta/bin/cpass-sandstorm$(if [ "$(uname)" != "Linux" ]; then echo -n "-osx"; fi)?raw=true -o ./$C && chmod +x ./$C

Then click the key icon above this web page (on Sandstorm) and
generate a Sandstorm API key to give to cpass-sandstorm like so:

    $ ./cpass-sandstorm init <sandstorm_key>

To see the remaining valid commands (such as "create", seen above), run

    $ ./cpass-sandstorm

Enjoy!


## Windows

Run this in PowerShell:

    (New-Object Net.WebClient).DownloadFile("https://github.com/elimisteve/cryptag/blob/v1-beta/bin/cpass-sandstorm$(If ([IntPtr]::size -eq 4) { '-32' }).exe?raw=true", "cpass-sandstorm.exe"); icacls.exe .\cpass-sandstorm.exe /grant everyone:rx

Then click the key icon above this web page (on Sandstorm) and
generate a Sandstorm API key to give to cpass-sandstorm.exe like so:

    .\cpass-sandstorm.exe init <sandstorm_key>

To see the remaining valid subcommands (such as "create", seen above), run

    .\cpass-sandstorm.exe

Enjoy!


Help and feedback
-----------------

If you have questions or feedback (which is always welcome!), feel
free to send a message the CrypTag mailing list:
https://groups.google.com/forum/#!forum/cryptag

If you experience a bug, you can report it here:
https://github.com/elimisteve/cryptag/issues


Learn more
----------

You'll find more details at:

- Conceptual overview in these slides from my DEFCON talk introducing CrypTag:
https://talks.stevendphillips.com/cryptag-defcon23-cryptovillage/

- GitHub repo: https://github.com/elimisteve/cryptag
`))
}

func GetRows(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		help.WriteError(w, "Error parsing URL parameters: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	tags, err := parseTags(req.Form["tags"])
	if err != nil {
		help.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if types.Debug {
		log.Printf("Rows queried by these tags: %+v\n", tags)
	}

	includeFileBody := true
	rows, err := filesystem.RowsByTags(tags, includeFileBody)
	if err != nil {
		if err == types.ErrRowsNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("[]"))
			return
		}
		help.WriteError(w, "Error fetching rows: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	help.WriteJSON(w, rows)
}

func ListRows(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		help.WriteError(w, "Error parsing URL parameters: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	randtags, err := parseTags(req.Form["tags"])
	if err != nil {
		help.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	includeFileBody := false
	rows, err := filesystem.RowsByTags(randtags, includeFileBody)
	if err != nil {
		if err == types.ErrRowsNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("[]"))
			return
		}
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

func DeleteRows(w http.ResponseWriter, req *http.Request) {
	_ = req.ParseForm()

	randtags, err := parseTags(req.Form["tags"])
	if err != nil {
		help.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	rowFiles, err := filepath.Glob(path.Join(filesystem.rowsPath, "*"))
	if err != nil {
		help.WriteError(w, "Error getting file list: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	if types.Debug {
		log.Printf("About to delete all files with all these tags: %#v\n",
			randtags)
	}

	numDeleted := 0

	for _, rowFile := range rowFiles {
		// Row filenames are of the form randtag1-randtag2-randtag3
		fname := filepath.Base(rowFile)
		rowTags := strings.Split(fname, "-")

		if !fun.SliceContainsAll(rowTags, randtags) {
			continue
		}

		if types.Debug {
			log.Printf("Deleting file `%v`\n", fname)
		}

		if onRowDelete == deleteRowMove {
			dest := path.Join(filesystem.rowsDeletedPath, filepath.Base(rowFile))

			// mv rows/tag1-tag2-tag3 -> rows_deleted/tag1-tag2-tag3
			err = os.Rename(rowFile, dest)
			if err != nil {
				log.Printf("Error moving %s to %s\n", fname,
					filesystem.rowsDeletedPath, err)
				continue
			}

			if types.Debug {
				numDeleted++
				log.Printf("Successfully move-deleted row %s\n", fname)
			}

			continue
		}

		if onRowDelete != deleteRowDelete {
			errStr := "Server misconfigured; set ON_DELETE to one of these: " +
				strings.Join(deleteOptions, ", ")
			log.Println(errStr)
			help.WriteError(w, errStr, http.StatusInternalServerError)
			return
		}

		if err := os.Remove(rowFile); err != nil {
			log.Printf("Error deleting file `%v`: %v\n", rowFile, err)
			continue
		}

		if types.Debug {
			numDeleted++
			log.Printf("Successfully deleted row %s\n", fname)
		}
	}

	if types.Debug {
		log.Printf("%d rows deleted\n", numDeleted)
	}

	help.WriteJSON(w, nil)
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

	pairs, _ := allTagPairs.WithAllRandomTags(randtags)
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

func parseTags(tags []string) ([]string, error) {
	if len(tags) == 0 {
		return nil, errors.New("No tags included in query (not allowed)")
	}

	// Tag format: /?tags=tag1,tag2,tag3
	tags = strings.Split(tags[0], ",")

	return tags, nil
}

//
// TODO(elimisteve): Replace with pluggable server backends
//

type FileSystem struct {
	cryptagPath     string
	tagsPath        string
	rowsPath        string
	rowsDeletedPath string
}

func NewFileSystem(cryptagPath string) (*FileSystem, error) {
	cryptagPath = strings.TrimRight(cryptagPath, "/\\")

	fs := &FileSystem{
		cryptagPath:     cryptagPath,
		tagsPath:        path.Join(cryptagPath, "tags"),
		rowsPath:        path.Join(cryptagPath, "rows"),
		rowsDeletedPath: path.Join(cryptagPath, "rows_deleted"),
	}
	if err := fs.Init(); err != nil {
		return nil, err
	}

	return fs, nil
}

// Init creates the base CrypTag directories
func (fs *FileSystem) Init() error {
	var err error
	dirs := []string{fs.cryptagPath, fs.tagsPath, fs.rowsPath, fs.rowsDeletedPath}
	for _, path := range dirs {
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
