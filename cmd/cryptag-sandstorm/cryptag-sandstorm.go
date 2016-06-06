// Steve Phillips / elimisteve
// 2015.02.24

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
	"github.com/elimisteve/cryptag/cli"
	"github.com/elimisteve/cryptag/cli/color"
	"github.com/elimisteve/cryptag/keyutil"
	"github.com/elimisteve/cryptag/rowutil"
)

var backendName = "sandstorm-webserver"

func init() {
	if bn := os.Getenv("BACKEND"); bn != "" {
		backendName = bn
	}
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln(usage)
	}

	var db *backend.WebserverBackend

	if os.Args[1] != "init" {
		var err error
		db, err = backend.LoadWebserverBackend("", backendName)
		if err != nil {
			log.Printf("%v\n", err)
			log.Fatal(usage)
		}

		if cryptag.UseTor {
			err := db.UseTor()
			if err != nil {
				log.Fatalf("Error trying to use Tor: %v\n", err)
			}
		}
	}

	switch os.Args[1] {
	case "init":
		if len(os.Args) < 3 {
			log.Fatalf("%s\n%s\n", initUsage, usage)
		}
		webkey := os.Args[2]
		if err := cli.InitSandstorm(backendName, webkey); err != nil {
			log.Fatal(err)
		}

	case "create", "createfile":
		if len(os.Args) < 4 {
			log.Println("At least 3 command line arguments must be included")
			log.Fatal(usage)
		}

		createFile := (os.Args[1] == "createfile")

		tags := append(os.Args[3:], "app:cryptag")

		if createFile {
			filename := os.Args[2]

			row, err := backend.CreateFileRow(db, nil, filename, tags)
			if err != nil {
				log.Fatalf("Error creating then saving file: %v", err)
			}

			color.Printf("Successfully saved new file %s with these tags:"+
				"\n%v\n", filepath.Base(filename), color.Tags(row.PlainTags()))
			return
		}

		// Create text row

		text := os.Args[2]
		tags = append(tags, "type:text")

		row, err := backend.CreateRow(db, nil, []byte(text), tags)
		if err != nil {
			log.Fatalf("Error creating text: %v\n", err)
		}

		color.Println(color.TextRow(row))

	case "getkey":
		fmt.Println(keyutil.Format(db.Key()))

	case "setkey":
		if len(os.Args) < 3 {
			log.Println("At least 2 command line arguments must be included")
			log.Fatal(usage)
		}

		keyStr := strings.Join(os.Args[2:], ",")

		err := backend.UpdateKey(db, keyStr)
		if err != nil {
			log.Fatalf("Error updating config with new key: %v", err)
		}

	case "list":
		plaintags := os.Args[2:]
		if len(plaintags) == 0 {
			plaintags = []string{"all"}
		}

		rows, err := backend.ListRowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		rowStrs := make([]string, len(rows))

		for i := range rows {
			// For non-file Rows, this will be empty string
			fname := rowutil.TagWithPrefixStripped(rows[i], "filename:")
			rowStrs[i] = color.TextAndTags(fname, rows[i].PlainTags())
		}
		color.Println(strings.Join(rowStrs, "\n\n"))

	case "get", "getfiles":
		getFile := (os.Args[1] == "getfiles")

		plaintags := os.Args[2:]
		if getFile {
			plaintags = append(plaintags, "type:file")
		} else {
			plaintags = append(plaintags, "type:text")
		}

		rows, err := backend.RowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		// Text rows; print then exit
		if !getFile {
			color.Println(color.TextRows(rows))
			return
		}

		// File rows; save them to files

		dir := path.Join(cryptag.TrustedBasePath, "decrypted", backendName)
		for _, row := range rows {
			fname, err := rowutil.SaveAsFile(row, dir)
			if err != nil {
				log.Printf("Error locally saving file: %s\n", err)
				continue
			}
			log.Printf("Successfully saved row to file %s", fname)
		}

	case "tags":
		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		for _, pair := range pairs {
			color.Printf("%s  %s\n", pair.Random, color.BlackOnWhite(pair.Plain()))
		}

	case "delete", "deletefiles":
		if len(os.Args) < 3 {
			log.Println("At least 2 command line arguments must be included")
			log.Fatal(usage)
		}

		deleteFiles := (os.Args[1] == "deletefiles")

		plaintags := os.Args[2:]
		if deleteFiles {
			plaintags = append(plaintags, "type:file")
		} else {
			plaintags = append(plaintags, "type:text")
		}

		if err := backend.DeleteRows(db, nil, plaintags); err != nil {
			log.Fatalf("Error deleting rows: %v\n", err)
		}

		log.Println("Row(s) successfully deleted")

	default:
		log.Printf("Subcommand `%s` not valid\n", os.Args[1])
		log.Println(usage)
	}
}

var usage = "Usage: " + filepath.Base(os.Args[0]) + " [create <yourpassword>] tag1 [tag2 ...]"
var initUsage = "Usage: " + filepath.Base(os.Args[0]) + " init <sandstorm_key>"
