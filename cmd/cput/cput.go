// Steve Phillips / elimisteve
// 2015.12.23

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/cli/color"
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
		log.Fatalln(createUsage)
	}

	switch os.Args[1] {
	default: // Create file
		if len(os.Args) < 3 {
			log.Printf("At least 2 command line arguments must be included\n")
			log.Fatalf(createUsage)
		}

		filename := os.Args[1]

		// TODO: Do streaming file reads
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Error reading file `%v`: %v\n", filename, err)
		}

		// Add tags
		tags := append(os.Args[2:], "app:cput", "type:file",
			"filename:"+filepath.Base(filename))
		lastDot := strings.LastIndex(filename, ".")
		if lastDot != -1 {
			fileExt := filename[lastDot+1:]
			tags = append(tags, "type:"+fileExt)
		}

		if types.Debug {
			log.Printf("Creating row with data of length `%v` and tags `%#v`\n",
				len(data), tags)
		}

		row, err := types.NewRow(data, tags)
		if err != nil {
			log.Fatalf("Error creating new row: %v\n", err)
		}

		err = backend.PopulateRowBeforeSave(db, row)
		if err != nil {
			log.Fatalf("Error populating row: %v\n", err)
		}

		err = db.SaveRow(row)
		if err != nil {
			log.Fatalf("Error saving new row: %v\n", err)
		}

		fmt.Printf("Successfully saved new row with these tags:\n\n%v\n",
			color.Tags(row.PlainTags()))
	}
}

var (
	createUsage = "Usage: " + filepath.Base(os.Args[0]) + " filename tag1 [tag2 ...]"
)
