// Steve Phillips / elimisteve
// 2016.01.19

package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/elimisteve/clipboard"
	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/cli/color"
	"github.com/elimisteve/cryptag/types"
)

var (
	db *backend.DropboxRemote
)

func init() {
	var err error
	db, err = backend.LoadDropboxRemote(
		os.Getenv("CRYPTAG_BACKEND_PATH"),
		os.Getenv("CRYPTAG_BACKEND_NAME"),
	)
	if err != nil {
		log.Fatalf("LoadDropboxRemote error: %v\n", err)
	}
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

	case "delete":
		if len(os.Args) < 3 {
			log.Printf("At least 2 command line arguments must be included\n")
			log.Fatalf(deleteUsage)
		}

		plaintags := append(os.Args[2:], "type:text")

		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		err = backend.DeleteRows(db, plaintags, pairs)
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

		// Ensures len(rows) > 0
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
