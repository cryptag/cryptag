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
	"github.com/elimisteve/cryptag/cli"
	"github.com/elimisteve/cryptag/cli/color"
	"github.com/elimisteve/cryptag/rowutil"
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
		cli.ArgFatal(allUsage)
	}

	switch os.Args[1] {
	// TODO: Add "decrypt" case
	// TODO: Add "file" case

	case "list":
		plaintags := append(os.Args[2:], "type:file")

		rows, err := backend.ListRowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		rowStrs := make([]string, len(rows))

		for i := range rows {
			fname := rowutil.TagWithPrefixStripped(rows[i], "filename:")
			rowStrs[i] = color.TextAndTags(fname, rows[i].PlainTags())
		}
		color.Println(strings.Join(rowStrs, "\n\n"))

	case "tags":
		pairs, err := db.AllTagPairs(nil)
		if err != nil {
			log.Fatal(err)
		}

		for _, pair := range pairs {
			color.Printf("%s  %s\n", pair.Random, color.BlackOnWhite(pair.Plain()))
		}

	default: // Decrypt, save to ~/.cryptag/decrypted/(filename from filename:...)
		plaintags := append(os.Args[1:], "type:file")

		rows, err := backend.RowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range rows {
			dir := path.Join(cryptag.TrustedBasePath, "decrypted")
			if _, err = rowutil.SaveAsFile(r, dir); err != nil {
				log.Printf("Error locally saving file: %v\n", err)
				continue
			}
			color.Printf("Successfully saved new row with these tags:\n%v\n",
				color.Tags(r.PlainTags()))
		}
	}
}

var (
	prefix = "Usage: " + filepath.Base(os.Args[0]) + " "

	getUsage  = prefix + "<tag1> [<tag2> ...]"
	listUsage = prefix + "list [<tag1> ...]"
	tagUsage  = prefix + "tags"

	allUsage = strings.Join([]string{getUsage, listUsage, tagUsage}, "\n")
)
