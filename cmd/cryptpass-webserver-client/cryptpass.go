// Steve Phillips / elimisteve
// 2015.02.24

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/elimisteve/clipboard"
	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/types"
)

var (
	SERVER_BASE_URL = "http://localhost:7777"
	SHARED_SECRET   = []byte(nil)
	AUTH_TOKEN      = ""

	db backend.Backend
)

func init() {
	types.Debug = false

	backend, err := backend.NewWebserverBackend(SHARED_SECRET, "webserver", SERVER_BASE_URL,
		AUTH_TOKEN)
	if err != nil {
		log.Fatalf("NewWebserverBackend error: %v\n", err)
	}
	db = backend
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln(usage)
	}

	switch os.Args[1] {
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

		newRow, err := types.NewRow([]byte(data), tags)
		if err != nil {
			log.Fatalf("Error creating new row: %v\n", err)
		}

		row, err := db.SaveRow(newRow)
		if err != nil {
			log.Fatalf("Error saving new row: %v\n", err)
		}
		fmt.Print(row.Format())

	default: // Search
		// Empty clipboard
		clipboard.WriteAll(nil)

		plaintags := append(os.Args[1:], "type:text")
		rows, err := db.RowsFromPlainTags(plaintags)
		if err != nil {
			log.Fatal(err)
		}

		if len(rows) == 0 {
			log.Fatal(types.ErrRowsNotFound)
		}

		// Add first row's contents to clipboard
		dec := rows[0].Decrypted()
		if err = clipboard.WriteAll(dec); err != nil {
			log.Fatalf("Error writing first result to clipboard: %v\n", err)
		}
		log.Printf("Added first result `%s` to clipboard\n", dec)

		fmt.Print(rows.Format())
	}
}

var usage = "Usage: " + filepath.Base(os.Args[0]) + " [create <yourpassword>] tag1 [tag2 ...]"
