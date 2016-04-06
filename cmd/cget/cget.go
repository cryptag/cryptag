// Steve Phillips / elimisteve
// 2015.12.23

package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/elimisteve/cryptag"
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
	// TODO: Add "decrypt" case
	// TODO: Add "file" case

	case "list":
		plaintags := append(os.Args[2:], "type:file")

		rows, err := db.ListRows(plaintags)
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range rows {
			fmt.Printf("%v\t\t%v\n\n", types.RowTagWithPrefix(r, "filename:"),
				strings.Join(r.PlainTags(), "  "))
		}

	case "tags":
		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}
		for _, pair := range pairs {
			fmt.Printf("%s  %s\n", pair.Random, pair.Plain())
		}

	default: // Decrypt, save to ~/.cryptag/decrypted/(filename from filename:...)
		plaintags := os.Args[1:]

		// TODO: Temporary?
		plaintags = append(plaintags, "type:file")

		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		rows, err := backend.RowsFromPlainTags(db, plaintags, pairs)
		if err != nil {
			log.Fatal(err)
		}

		if len(rows) == 0 {
			log.Fatal(types.ErrRowsNotFound)
		}

		var rowFilename string
		for _, r := range rows {
			dir := path.Join(cryptag.TrustedBasePath, "decrypted")
			if rowFilename, err = types.SaveRowAsFile(r, dir); err != nil {
				log.Printf("Error locally saving file: %v\n", err)
				continue
			}
			log.Printf("Saved new file: %v\n", rowFilename)
		}
	}
}

var (
	usage = "Usage: " + filepath.Base(os.Args[0]) + " tag1 [tag2 ...]"
)
