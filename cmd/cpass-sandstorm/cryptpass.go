// Steve Phillips / elimisteve
// 2015.02.24

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/elimisteve/clipboard"
	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/cli/color"
	"github.com/elimisteve/cryptag/types"
)

var backendName = "sandstorm-webserver"

func init() {
	if bn := os.Getenv("CRYPTAG_BACKEND_NAME"); bn != "" {
		backendName = bn
	}
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln(usage)
	}

	var db backend.Backend

	if os.Args[1] != "init" {
		var err error
		db, err = backend.LoadWebserverBackend("", backendName)
		if err != nil {
			log.Printf("%v\n", err)
			log.Fatal(usage)
		}
	}

	switch os.Args[1] {
	case "init":
		if len(os.Args) < 3 {
			log.Fatal(initUsage)
		}
		if err := createBackendConfig(os.Args[2]); err != nil {
			log.Fatal(err)
		}
	case "create":
		if len(os.Args) < 4 {
			log.Println("At least 3 command line arguments must be included")
			log.Fatal(usage)
		}

		data := os.Args[2]
		tags := append(os.Args[3:], "app:cryptpass", "type:text")

		if types.Debug {
			log.Printf("Creating row with data `%s` and tags `%#v`\n", data, tags)
		}

		row, err := types.NewRow([]byte(data), tags)
		if err != nil {
			log.Fatalf("Error creating new row: %v\n", err)
		}

		err = db.SaveRow(row)
		if err != nil {
			log.Fatalf("Error saving new row: %v\n", err)
		}
		color.Println(color.TextRow(row))

	default: // Search
		// Empty clipboard
		clipboard.WriteAll(nil)

		plaintags := append(os.Args[1:], "type:text")

		// TODO: Consider caching locally
		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		rows, err := backend.RowsFromPlainTags(db, plaintags, pairs)
		if err != nil {
			log.Fatal(err)
		}

		// Add first row's contents to clipboard
		dec := rows[0].Decrypted()
		if err = clipboard.WriteAll(dec); err != nil {
			log.Fatalf("Error writing first result to clipboard: %v\n", err)
		}
		log.Printf("Added first result `%s` to clipboard\n", dec)

		color.Println(color.TextRows(rows))
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
