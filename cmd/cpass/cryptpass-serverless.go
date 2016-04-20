// Steve Phillips / elimisteve
// 2015.11.04

package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/elimisteve/clipboard"
	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/cli/color"
)

var (
	db backend.Backend
)

func init() {
	fs, err := backend.LoadOrCreateFileSystem(
		os.Getenv("BACKEND_PATH"),
		os.Getenv("BACKEND"),
	)
	if err != nil {
		log.Fatalf("LoadOrCreateFileSystem error: %v\n", err)
	}

	db = fs
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln(usage)
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 4 {
			log.Printf("At least 3 command line arguments must be included\n")
			log.Fatalf(createUsage)
		}

		data := os.Args[2]
		tags := append(os.Args[3:], "app:cryptpass", "type:text")

		row, err := backend.CreateRow(db, nil, []byte(data), tags)
		if err != nil {
			log.Fatalf("Error creating then saving new row: %v", err)
		}

		color.Println(color.TextRow(row))

	case "delete":
		if len(os.Args) < 3 {
			log.Printf("At least 2 command line arguments must be included\n")
			log.Fatalf(deleteUsage)
		}
		plaintags := append(os.Args[2:], "type:text")

		err := backend.DeleteRows(db, plaintags, nil)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Row(s) successfully deleted\n")

	default: // Search
		// Empty clipboard
		clipboard.WriteAll(nil)

		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		plaintags := append(os.Args[1:], "type:text")
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

var (
	usage       = "Usage: " + filepath.Base(os.Args[0]) + " [create <yourpassword> | delete] tag1 [tag2 ...]"
	createUsage = "Usage: " + filepath.Base(os.Args[0]) + " create <yourpassword> tag1 [tag2 ...]"
	deleteUsage = "Usage: " + filepath.Base(os.Args[0]) + " delete tag1 [tag2 ...]"
)
