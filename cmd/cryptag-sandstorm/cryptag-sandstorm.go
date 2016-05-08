// Steve Phillips / elimisteve
// 2015.02.24

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/cli/color"
	"github.com/elimisteve/cryptag/types"
)

var backendName = "sandstorm-webserver"

func init() {
	if bn := os.Getenv("BACKEND"); bn != "" {
		backendName = bn
	}
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln(usage)
	}

	var db *backend.WebserverBackend

	if os.Args[1] != "init" {
		var err error
		db, err = backend.LoadWebserverBackend("", backendName)
		if err != nil {
			log.Printf("%v\n", err)
			log.Fatal(usage)
		}
	}

	if cryptag.UseTor {
		err := db.UseTor()
		if err != nil {
			log.Fatalf("Error trying to use Tor: %v\n", err)
		}
	}

	switch os.Args[1] {
	case "init":
		if len(os.Args) < 3 {
			log.Fatalf("%s\n%s\n", initUsage, usage)
		}
		if err := createBackendConfig(os.Args[2]); err != nil {
			log.Fatal(err)
		}

	case "create", "createfile":
		if len(os.Args) < 4 {
			log.Println("At least 3 command line arguments must be included")
			log.Fatal(usage)
		}

		createFile := (os.Args[1] == "createfile")

		var rowData []byte
		tags := append(os.Args[3:], "app:cryptag")

		if createFile {
			filename := os.Args[2]

			b, err := ioutil.ReadFile(filename)
			if err != nil {
				log.Fatalf("Error reading file `%s`: %v\n", filename, err)
			}

			rowData = b
			tags = append(tags, "type:file", "filename:"+filepath.Base(filename))
		} else {
			rowData = []byte(os.Args[2])
			tags = append(tags, "type:text")
		}

		row, err := backend.CreateRow(db, nil, rowData, tags)
		if err != nil {
			log.Fatalf("Error creating then saving new row: %v", err)
		}

		if createFile {
			color.Printf("Successfully saved new row with these tags:\n%v\n",
				color.Tags(row.PlainTags()))
		} else {
			color.Println(color.TextRow(row))
		}

	case "getkey":
		fmt.Println(fmtKey(db.Key()))

	case "setkey":
		if len(os.Args) < 3 {
			log.Println("At least 2 command line arguments must be included")
			log.Fatal(usage)
		}

		keyStr := strings.Join(os.Args[2:], ",")

		newKey, err := parseKey(keyStr)
		if err != nil {
			log.Fatalf("Error from parseKey: %v\n", err)
		}

		cfg, err := db.ToConfig()
		if err != nil {
			log.Fatal(err)
		}

		cfg.Key = newKey

		if err := cfg.Update(cryptag.BackendPath); err != nil {
			log.Fatalf("Error updating config: %v", err)
		}

	case "list":
		plaintags := os.Args[2:]
		if len(plaintags) == 0 {
			plaintags = []string{"all"}
		}

		rows, err := backend.ListRowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		rowStrs := make([]string, len(rows))

		for i := range rows {
			// For non-file Rows, this will be empty string
			fname := types.RowTagWithPrefix(rows[i], "filename:")
			rowStrs[i] = color.TextAndTags(fname, rows[i].PlainTags())
		}
		color.Println(strings.Join(rowStrs, "\n\n"))

	case "get", "getfiles":
		getFile := (os.Args[1] == "getfiles")

		plaintags := os.Args[2:]
		if getFile {
			plaintags = append(plaintags, "type:file")
		} else {
			plaintags = append(plaintags, "type:text")
		}

		rows, err := backend.RowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		// Text rows; print then exit
		if !getFile {
			color.Println(color.TextRows(rows))
			return
		}

		// File rows; save them to files

		dir := path.Join(cryptag.TrustedBasePath, "decrypted", backendName)
		for _, row := range rows {
			fname, err := types.SaveRowAsFile(row, dir)
			if err != nil {
				log.Printf("Error locally saving file: %s\n", err)
				continue
			}
			log.Printf("Successfully saved row to file %s", fname)
		}

	case "tags":
		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		for _, pair := range pairs {
			color.Printf("%s  %s\n", pair.Random, color.BlackOnWhite(pair.Plain()))
		}

	case "delete", "deletefiles":
		if len(os.Args) < 3 {
			log.Println("At least 2 command line arguments must be included")
			log.Fatal(usage)
		}

		deleteFiles := (os.Args[1] == "deletefiles")

		plaintags := os.Args[2:]
		if deleteFiles {
			plaintags = append(plaintags, "type:file")
		} else {
			plaintags = append(plaintags, "type:text")
		}

		if err := backend.DeleteRows(db, nil, plaintags); err != nil {
			log.Fatalf("Error deleting rows: %v\n", err)
		}

		log.Println("Row(s) successfully deleted")

	default:
		log.Printf("Subcommand `%s` not valid\n", os.Args[1])
		log.Println(usage)
	}
}

var usage = "Usage: " + filepath.Base(os.Args[0]) + " [create <yourpassword>] tag1 [tag2 ...]"
var initUsage = "Usage: " + filepath.Base(os.Args[0]) + " init <sandstorm_key>"

func createBackendConfig(key string) error {
	info := strings.SplitN(key, "#", 2)
	if len(info) < 2 {
		return fmt.Errorf(
			"Error parsing `%v` as Sandstorm key generated from Sandstorm's UI",
			key)
	}

	serverBaseURL, authToken := info[0], info[1]

	db, err := backend.NewWebserverBackend(nil, backendName, serverBaseURL, authToken)
	if err != nil {
		return fmt.Errorf("NewWebserverBackend error: %v\n", err)
	}

	cfg, err := db.ToConfig()
	if err != nil {
		return fmt.Errorf("Error getting backend config: %v\n", err)
	}

	err = cfg.Save(cryptag.BackendPath)
	if err != nil && err != backend.ErrConfigExists {
		return fmt.Errorf("Error saving backend config to disk: %v\n", err)
	}

	return nil
}

var keyRegex = regexp.MustCompile(`(\d+)`)

func parseKey(cliDigits string) (*[32]byte, error) {
	// Pluck out all digit sequences, convert to numbers
	nums := keyRegex.FindAllString(cliDigits, -1)
	if len(nums) != 32 {
		return nil, fmt.Errorf("Key must include 32 numbers, not %d", len(nums))
	}

	var newKey [32]byte

	for i := 0; i < 32; i++ {
		n, err := strconv.ParseUint(nums[i], 10, 8)
		if err != nil {
			return nil, fmt.Errorf("Number #%d '%v' was invalid: %v\n", i+1,
				nums[i])
		}
		newKey[i] = byte(n)
	}

	return &newKey, nil
}

func fmtKey(key *[32]byte) string {
	if key == nil {
		return "<nil>"
	}
	k := *key

	kStr := fmt.Sprintf("%d", k[0])

	for i := 1; i < len(k); i++ {
		kStr += fmt.Sprintf(",%d", k[i])
	}
	return kStr
}
