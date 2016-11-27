// Steve Phillips / elimisteve
// 2015.11.04

package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cryptag/cryptag/backend"
	"github.com/cryptag/cryptag/cli"
	"github.com/cryptag/cryptag/cli/color"
	"github.com/cryptag/cryptag/importer"
	"github.com/cryptag/cryptag/rowutil"
	"github.com/elimisteve/clipboard"
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
		cli.ArgFatal(allUsage)
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 4 {
			cli.ArgFatal(createUsage)
		}

		data := os.Args[2]
		tags := append(os.Args[3:], "app:cryptpass", "type:text")

		row, err := backend.CreateRow(db, nil, []byte(data), tags)
		if err != nil {
			log.Fatalf("Error creating then saving new row: %v", err)
		}

		color.Println(color.TextRow(row))

	case "tags":
		pairs, err := db.AllTagPairs(nil)
		if err != nil {
			log.Fatal(err)
		}

		for _, pair := range pairs {
			color.Printf("%s  %s\n", pair.Random, color.BlackOnWhite(pair.Plain()))
		}

	case "delete":
		if len(os.Args) < 3 {
			cli.ArgFatal(deleteUsage)
		}

		plaintags := append(os.Args[2:], "type:text")

		err := backend.DeleteRows(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Row(s) successfully deleted\n")

	case "import":
		if len(os.Args) < 3 {
			cli.ArgFatal(importUsage)
		}

		filename := os.Args[2]
		plaintags := os.Args[3:]

		rows, err := importer.KeePassCSV(filename, plaintags)
		if err != nil {
			log.Fatalf("Error importing KeePass CSV `%v`: %v", filename, err)
		}

		pairs, err := db.AllTagPairs(nil)
		if err != nil {
			log.Fatalf("Error fetching all TagPairs: %v\n", err)
		}

		for _, row := range rows {
			if _, err = backend.PopulateRowBeforeSave(db, row, pairs); err != nil {
				log.Printf("Error decrypting row %#v: %v\n", row, err)
				continue
			}
			if err := db.SaveRow(row); err != nil {
				log.Printf("Error saving row %#v: %v\n", row, err)
				continue
			}

			log.Printf("Successfully imported password for site %s\n",
				rowutil.TagWithPrefixStripped(row, "url:"))
		}

	default: // Search
		// Empty clipboard
		clipboard.WriteAll(nil)

		plaintags := append(os.Args[1:], "type:text")
		rows, err := backend.RowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		rows.Sort(rowutil.ByTagPrefix("created:", true))

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
	prefix = "Usage: " + filepath.Base(os.Args[0]) + " "

	createUsage = prefix + "create <password or text> <tag1> [type:password <tag3> ...]"
	tagsUsage   = prefix + "tags"
	deleteUsage = prefix + "delete <tag1> [<tag2> ...]"
	importUsage = prefix + "import <exported-from-keepassx.csv> [<tag1> ...]"

	allUsage = strings.Join([]string{createUsage, tagsUsage, deleteUsage, importUsage}, "\n")
)
