// Steve Phillips / elimisteve
// 2016.01.19

package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/elimisteve/clipboard"
	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/cli/color"
)

var (
	db *backend.DropboxRemote
)

func init() {
	var err error
	db, err = backend.LoadDropboxRemote(
		os.Getenv("BACKEND_PATH"),
		os.Getenv("BACKEND"),
	)
	if err != nil {
		log.Fatalf("LoadDropboxRemote error: %v\n", err)
	}

	if cryptag.UseTor {
		err = db.UseTor()
		if err != nil {
			log.Fatalf("Error creating Tor HTTP client: %v\n", err)
		}
	}
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln(usage)
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 4 {
			log.Println("At least 3 command line arguments must be included")
			log.Fatal(createUsage)
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
			log.Println("At least 2 command line arguments must be included")
			log.Fatal(deleteUsage)
		}

		plaintags := append(os.Args[2:], "type:text")

		err := backend.DeleteRows(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Row(s) successfully deleted")

	default: // Search
		// Empty clipboard
		clipboard.WriteAll(nil)

		plaintags := append(os.Args[1:], "type:text")

		// Ensures len(rows) > 0
		rows, err := backend.RowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		// Add first row's contents to clipboard
		dec := rows[0].Decrypted()
		if err = clipboard.WriteAll(dec); err != nil {
			log.Printf("Error writing first result to clipboard: %v\n", err)
		} else {
			log.Printf("Added first result `%s` to clipboard\n", dec)
		}

		color.Println(color.TextRows(rows))
	}
}

var (
	usage       = "Usage: " + filepath.Base(os.Args[0]) + " [create <yourpassword> | delete] tag1 [tag2 ...]"
	createUsage = "Usage: " + filepath.Base(os.Args[0]) + " create <yourpassword> tag1 [tag2 ...]"
	deleteUsage = "Usage: " + filepath.Base(os.Args[0]) + " delete tag1 [tag2 ...]"
)
