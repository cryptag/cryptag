// Steve Phillips / elimisteve
// 2015.11.04

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
	db backend.Backend
)

func init() {
	fs, err := backend.LoadOrCreateFileSystem(
		os.Getenv("CRYPTAG_BACKEND_PATH"),
		os.Getenv("CRYPTAG_BACKEND_NAME"),
	)
	if err != nil {
		log.Fatalf("LoadFileSystem error: %v\n", err)
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
		fmt.Print(row.Format())

	case "delete":
		if len(os.Args) < 3 {
			log.Printf("At least 2 command line arguments must be included\n")
			log.Fatalf(deleteUsage)
		}
		plainTags := os.Args[2:]

		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatalf("Error from AllTagPairs: %v\n", err)
		}

		// Get all the random tags associated with the tag pairs that
		// contain every tag in plainTags.
		//
		// Got that?
		//
		// The flow: user specifies plainTags + we fetch all TagPairs
		// => we filter the TagPairs based on those with the
		// user-specified plainTags => we grab each TagPair's random
		// string so we can delete the rows tagged with those tags

		pairs, err = pairs.HaveAllPlainTags(plainTags)
		if err != nil {
			log.Fatal(err)
		}

		// Delete rows tagged with the random strings in pairs
		err = db.DeleteRows(pairs.AllRandom())
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

var (
	usage       = "Usage: " + filepath.Base(os.Args[0]) + " [create <yourpassword> | delete] tag1 [tag2 ...]"
	createUsage = "Usage: " + filepath.Base(os.Args[0]) + " create <yourpassword> tag1 [tag2 ...]"
	deleteUsage = "Usage: " + filepath.Base(os.Args[0]) + " delete tag1 [tag2 ...]"
)
