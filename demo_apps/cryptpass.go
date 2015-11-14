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
	SERVER_BASE_URL = ""
	SHARED_SECRET   = nil

	db backend.Backend
)

func init() {
	types.Debug = true

	backend, err := backend.NewWebserverBackend(SHARED_SECRET, SERVER_BASE_URL)
	if err != nil {
		log.Fatalf("NewWebserverBackend error: %v\n", err)
	}
	db = backend
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalf(usage)
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 4 {
			log.Printf("At least 3 command line arguments must be included\n")
			log.Fatalf(usage)
		}

		data := os.Args[2]
		tags := append(os.Args[3:], "app:cryptpass")

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

		plaintags := os.Args[1:]
		rows, err := db.RowsFromPlainTags(plaintags)
		if err != nil {
			log.Fatalf("Error from RowsFromPlainTags: %v\n", err)
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

var usage = "Usage: " + filepath.Base(os.Args[0]) + " data tag1 [tag2 ...]"
