// Steve Phillips / elimisteve
// 2015.12.23

package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/cryptag/cryptag/backend"
	"github.com/cryptag/cryptag/cli"
	"github.com/cryptag/cryptag/cli/color"
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
		cli.Fatal(createUsage)
	}

	switch os.Args[1] {
	default: // Create file
		if len(os.Args) < 3 {
			cli.ArgFatal(createUsage)
		}

		filename := os.Args[1]
		tags := append(os.Args[2:], "app:cput")

		row, err := backend.CreateFileRow(db, nil, filename, tags)
		if err != nil {
			log.Fatalf("Error saving file: %v\n", err)
		}

		color.Printf("Successfully saved new row with these tags:\n%v\n",
			color.Tags(row.PlainTags()))
	}
}

var (
	prefix = "Usage: " + filepath.Base(os.Args[0]) + " "

	createUsage = prefix + "<filename> <tag1> [<tag2> ...]"
)
