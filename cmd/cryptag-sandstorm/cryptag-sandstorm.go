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
		cli.Fatal(allUsage)
	}

	var db *backend.WebserverBackend

	if os.Args[1] != "init" {
		var err error
		db, err = backend.LoadWebserverBackend("", backendName)
		if err != nil {
			log.Fatalf("Error loading config for webserver backend `%s`: %v",
				backendName, err)
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
			cli.ArgFatal(initUsage)
		}

		webkey := os.Args[2]

		if err := cli.InitSandstorm(backendName, webkey); err != nil {
			log.Fatal(err)
		}

	case "createtext", "ct", "createfile", "cf", "createany", "ca":
		if len(os.Args) < 4 {
			cli.ArgFatal(allCreateUsage)
		}

		createFile := (os.Args[1] == "createfile" || os.Args[1] == "cf")
		createAny := (os.Args[1] == "createany" || os.Args[1] == "ca")

		tags := append(os.Args[3:], "app:cryptag")

		// Create file row
		if createFile {
			filename := os.Args[2]

			row, err := backend.CreateFileRow(db, nil, filename, tags)
			if err != nil {
				log.Fatalf("Error creating then saving file: %v", err)
			}

			color.Printf("%s successfully saved with these tags:"+
				"\n%v\n", color.BlackOnCyan(filepath.Base(filename)),
				color.Tags(row.PlainTags()))
			return
		}

		//
		// Create text row _or_ custom row
		//

		data := os.Args[2]
		if !createAny {
			tags = append(tags, "type:text")
		}

		row, err := backend.CreateRow(db, nil, []byte(data), tags)
		if err != nil {
			log.Fatalf("Error creating text: %v\n", err)
		}

		color.Println(color.TextRow(row))

	case "getkey":
		fmt.Println(keyutil.Format(db.Key()))

	case "setkey":
		if len(os.Args) < 3 {
			cli.ArgFatal(setkeyUsage)
		}

		keyStr := strings.Join(os.Args[2:], ",")

		err := backend.UpdateKey(db, keyStr)
		if err != nil {
			log.Fatalf("Error updating config with new key: %v", err)
		}

	case "listtext", "lt", "listfiles", "lf", "listany", "la":
		listFile := (os.Args[1] == "listfiles" || os.Args[1] == "lf")
		listAny := (os.Args[1] == "listany" || os.Args[1] == "la")

		plaintags := append(os.Args[2:], "all")

		if !listAny {
			if listFile {
				plaintags = append(plaintags, "type:file")
			} else {
				plaintags = append(plaintags, "type:text")
			}
		}

		rows, err := backend.ListRowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		rows.Sort(rowutil.ByTagPrefix("created:", true))

		rowStrs := make([]string, len(rows))

		for i := range rows {
			// For non-file Rows, this will be empty string
			fname := rowutil.TagWithPrefixStripped(rows[i], "filename:")
			rowStrs[i] = color.TextAndTags(fname, rows[i].PlainTags())
		}

		color.Println(strings.Join(rowStrs, "\n\n"))

	case "gettext", "gt", "getfiles", "gf", "getany", "ga":
		getFile := (os.Args[1] == "getfiles" || os.Args[1] == "gf")
		getAny := (os.Args[1] == "getany" || os.Args[1] == "ga")

		plaintags := append(os.Args[2:], "all")

		if !getAny {
			if getFile {
				plaintags = append(plaintags, "type:file")
			} else {
				plaintags = append(plaintags, "type:text")
			}
		}

		rows, err := backend.RowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		// Sort oldest to newest
		rows.Sort(rowutil.ByTagPrefix("created:", true))

		dir := path.Join(cryptag.TrustedBasePath, "decrypted", backendName)
		for i, row := range rows {
			if i != 0 {
				fmt.Println("")
			}

			// Print bodies of non-file rows as text (includes Tasks, etc)
			if !row.HasPlainTag("type:file") {
				color.Println(color.TextRow(row))
				continue
			}

			fname, err := rowutil.SaveAsFile(row, dir)
			if err != nil {
				log.Printf("Error locally saving file: %s\n", err)
				continue
			}
			color.Printf("%s successfully downloaded; has these tags:\n%v\n",
				color.BlackOnCyan(fname), color.Tags(row.PlainTags()))
		}

	case "tags", "t":
		pairs, err := db.AllTagPairs(nil)
		if err != nil {
			log.Fatal(err)
		}

		for _, pair := range pairs {
			color.Printf("%s  %s\n", pair.Random, color.BlackOnWhite(pair.Plain()))
		}

	case "deletetext", "dt", "deletefiles", "df", "deleteany", "da":
		if len(os.Args) < 3 {
			cli.ArgFatal(allDeleteUsage)
		}

		deleteFiles := (os.Args[1] == "deletefiles" || os.Args[1] == "df")
		deleteAny := (os.Args[1] == "deleteany" || os.Args[1] == "da")

		plaintags := append(os.Args[2:], "all")

		if !deleteAny {
			if deleteFiles {
				plaintags = append(plaintags, "type:file")
			} else {
				plaintags = append(plaintags, "type:text")
			}
		}

		if err := backend.DeleteRows(db, nil, plaintags); err != nil {
			log.Fatalf("Error deleting rows: %v\n", err)
		}

		log.Println("Row(s) successfully deleted")

	default:
		log.Printf("Subcommand `%s` not valid\n", os.Args[1])
		cli.Fatal(allUsage)
	}
}

var (
	prefix = "Usage: " + filepath.Base(os.Args[0]) + " "

	initUsage = prefix + "init <sandstorm_webkey>"

	createTextUsage = prefix + "createtext <text>     <tag1> [<tag2> ...]"
	createFileUsage = prefix + "createfile <filename> <tag1> [<tag2> ...]"
	createAnyUsage  = prefix + "createany  <data>     <tag1> [<type:...> <tag2> ...]"
	allCreateUsage  = strings.Join([]string{createTextUsage, createFileUsage, createAnyUsage}, "\n")

	getTextUsage  = prefix + "gettext  <tag1> [<tag2> ...]"
	getFilesUsage = prefix + "getfiles <tag1> [<tag2> ...]"
	getAnyUsage   = prefix + "getany   <tag1> [<tag2> ...]"
	allGetUsage   = strings.Join([]string{getTextUsage, getFilesUsage, getAnyUsage}, "\n")

	deleteTextUsage  = prefix + "deletetext  <tag1> [<tag2> ...]"
	deleteFilesUsage = prefix + "deletefiles <tag1> [<tag2> ...]"
	deleteAnyUsage   = prefix + "deleteany   <tag1> [<tag2> ...]"
	allDeleteUsage   = strings.Join([]string{deleteTextUsage, deleteFilesUsage, deleteAnyUsage}, "\n")

	getkeyUsage = prefix + "getkey"
	setkeyUsage = prefix + "setkey <key>"

	allUsages = []string{
		initUsage, "",
		createTextUsage, createFileUsage, createAnyUsage, "",
		getTextUsage, getFilesUsage, getAnyUsage, "",
		deleteTextUsage, deleteFilesUsage, deleteAnyUsage, "",
		getkeyUsage, setkeyUsage,
	}
	allUsage = strings.Join(allUsages, "\n")
)
