// Steve Phillips / elimisteve
// 2015.12.23

package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/cli/color"
	"github.com/elimisteve/cryptag/types"
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

		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		rows, err := backend.ListRowsFromPlainTags(db, pairs, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		rowStrs := make([]string, len(rows))

		for i := range rows {
			fname := types.RowTagWithPrefix(rows[i], "filename:")
			rowStrs[i] = color.TextAndTags(fname, rows[i].PlainTags())
		}
		color.Println(strings.Join(rowStrs, "\n\n"))

	case "tags":
		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}
		for _, pair := range pairs {
			color.Printf("%s  %s\n", pair.Random, color.BlackOnWhite(pair.Plain()))
		}

	default: // Decrypt, save to ~/.cryptag/decrypted/(filename from filename:...)
		plaintags := os.Args[1:]

		// TODO: Temporary?
		plaintags = append(plaintags, "type:file")

		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		rows, err := backend.RowsFromPlainTags(db, pairs, plaintags)
		if err != nil {
			log.Fatal(err)
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
