// Steve Phillips / elimisteve
// 2015.02.24

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/types"
)

var (
	SERVER_BASE_URL = ""
	SHARED_SECRET   = ""

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
		tags := os.Args[3:]

		newRow := types.NewRow([]byte(data), tags)

		row, err := db.SaveRow(newRow)
		if err != nil {
			log.Fatalf("Error saving new row: %v\n", err)
		}
		fmt.Print(row.Format())

	default: // Search
		plaintags := os.Args[1:]
		rows, err := db.RowsFromPlainTags(plaintags)
		if err != nil {
			log.Fatalf("Error from RowsFromPlainTags: %v\n", err)
		}
		if err = rows.FirstToClipboard(); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing first result to clipboard: %v\n", err)
		}
		fmt.Print(rows.Format())
	}
}

var usage = "Usage: " + os.Args[0] + " data tag1 [tag2 ...]"
