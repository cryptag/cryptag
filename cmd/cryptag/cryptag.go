// Steven Phillips / elimisteve
// 2016.08.11

package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/backend"
	"github.com/cryptag/cryptag/cli"
	"github.com/cryptag/cryptag/cli/color"
	"github.com/cryptag/cryptag/keyutil"
	"github.com/cryptag/cryptag/rowutil"
)

var (
	backendName = os.Getenv("BACKEND")
	dashB       = false
)

func init() {
	if len(os.Args) >= 4 && os.Args[1] == "-b" {
		backendName = os.Args[2]
		dashB = true
	}
	if backendName == "" {
		backendName = "default"
	}
}

func main() {
	if len(os.Args) == 1 {
		cli.Fatal(allUsage)
	}

	var db backend.Backend

	osArgs := os.Args
	if dashB {
		// Skip over '-b myBackendName'
		osArgs = append([]string{os.Args[0]}, os.Args[3:]...)
	}

	if !containsAny(osArgs[1], "init", "listbackends", "lb",
		"setdefaultbackend", "sdb") {

		var err error
		db, err = backend.LoadBackend("", backendName)
		if err != nil {
			log.Fatalf("Error loading config for backend `%s`: %v",
				backendName, err)
		}

		if db, ok := db.(cryptag.CanUseTor); ok && cryptag.UseTor {
			err = db.UseTor()
			if err != nil {
				log.Fatalf("Error trying to use Tor: %v\n", err)
			}
		}
	}

	switch osArgs[1] {
	case "init":
		if len(osArgs) < 5 {
			cli.ArgFatal(allInitUsage)
		}

		backendType := osArgs[2]
		backendName := osArgs[3]
		args := osArgs[4:]

		if backendType == "sandstorm" {
			backendType = backend.TypeWebserver
		}

		if _, err := backend.Create(backendType, backendName, args); err != nil {
			log.Fatal(err)
		}

	case "createtext", "ct", "createfile", "cf", "createany", "ca":
		if len(osArgs) < 4 {
			cli.ArgFatal(allCreateUsage)
		}

		createFile := (osArgs[1] == "createfile" || osArgs[1] == "cf")
		createAny := (osArgs[1] == "createany" || osArgs[1] == "ca")

		tags := append(osArgs[3:], "app:cryptag")

		// Create file row
		if createFile {
			filename := osArgs[2]

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

		data := osArgs[2]
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
		if len(osArgs) < 3 {
			cli.ArgFatal(setkeyUsage)
		}

		keyStr := strings.Join(osArgs[2:], ",")

		err := backend.UpdateKey(db, keyStr)
		if err != nil {
			log.Fatalf("Error updating config with new key: %v", err)
		}

	case "listbackends", "lb":
		bkPattern := "*"
		typ := ""

		if len(osArgs) > 2 {
			if strings.HasPrefix(osArgs[2], "type:") {
				typ = strings.TrimPrefix(osArgs[2], "type:")
			} else {
				bkPattern = osArgs[2]
			}
		}

		configs, err := backend.ReadConfigs("", bkPattern)
		if err != nil {
			log.Printf("Error reading Backend configs: %v\n", err)

			// Fall through
		}

		for _, conf := range configs {
			// If user specified type, skip over configs of the wrong
			// type
			if typ != "" && typ != conf.GetType() {
				continue
			}

			current := ""
			if conf.Name == backendName {
				current = "*"
			}

			color.Printf("%-40s   %-30s   %s\n",
				current+color.BlackOnCyan(conf.Name),
				color.BlackOnWhite(conf.GetType()),
				color.BlackOnWhite(conf.GetPath()),
			)
		}

	case "setdefaultbackend", "sdb":
		if len(osArgs) < 3 {
			cli.ArgFatal(setDefaultBackendUsage)
		}

		if err := backend.SetDefaultBackend(db, "", osArgs[2]); err != nil {
			log.Fatal(err)
		}

	case "listtext", "lt", "listfiles", "lf", "listany", "la":
		listFiles := (osArgs[1] == "listfiles" || osArgs[1] == "lf")
		listAny := (osArgs[1] == "listany" || osArgs[1] == "la")

		plaintags := append(osArgs[2:], "all")

		if !listAny {
			if listFiles {
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
		getFiles := (osArgs[1] == "getfiles" || osArgs[1] == "gf")
		getAny := (osArgs[1] == "getany" || osArgs[1] == "ga")

		plaintags := append(osArgs[2:], "all")

		if !getAny {
			if getFiles {
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
			if !getFiles {
				color.Println(color.TextRow(row))
				continue
			}

			fname, err := rowutil.SaveAsFile(row, dir)
			if err != nil {
				log.Printf("Error locally saving file: %s\n", err)
				continue
			}
			color.Printf("%s successfully saved; has these tags:\n%v\n",
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
		if len(osArgs) < 3 {
			cli.ArgFatal(allDeleteUsage)
		}

		deleteFiles := (osArgs[1] == "deletefiles" || osArgs[1] == "df")
		deleteAny := (osArgs[1] == "deleteany" || osArgs[1] == "da")

		plaintags := append(osArgs[2:], "all")

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
		log.Printf("Subcommand `%s` not valid\n", osArgs[1])
		cli.Fatal(allUsage)
	}
}

func containsAny(in string, strs ...string) bool {
	for _, s := range strs {
		if in == s {
			return true
		}
	}
	return false
}

var (
	prefix = "Usage: " + filepath.Base(os.Args[0]) + " "

	initFilesystemUsage = prefix + "init filesystem <backend name> <data base path>"
	initSandstormUsage  = prefix + "init sandstorm  <backend name> <sandstorm web key>"
	initWebserverUsage  = prefix + "init webserver  <backend name> <base url> <auth token>"
	initDropboxUsage    = prefix + "init dropbox    <backend name> <app key> <app secret> <access token> <base path>"
	allInitUsage        = strings.Join([]string{initFilesystemUsage,
		initSandstormUsage, initWebserverUsage, initDropboxUsage}, "\n")

	createTextUsage = prefix + "createtext <text>     <tag1> [<tag2> ...]"
	createFileUsage = prefix + "createfile <filename> <tag1> [<tag2> ...]"
	createAnyUsage  = prefix + "createany  <data>     <tag1> [<tag2> <type:...> ...]"
	allCreateUsage  = strings.Join([]string{createTextUsage, createFileUsage, createAnyUsage}, "\n")

	listBackendsUsage = prefix + "listbackends [ <name-matching regex> | type:(dropbox|filesystem|webserver) ]"

	setDefaultBackendUsage = prefix + "setdefaultbackend <backend name>"

	listTextUsage  = prefix + "listtext  <tag1> [<tag2> ...]"
	listFilesUsage = prefix + "listfiles <tag1> [<tag2> ...]"
	listAnyUsage   = prefix + "listany   <tag1> [<tag2> ...]"
	allListUsage   = strings.Join([]string{listTextUsage, listFilesUsage, listAnyUsage}, "\n")

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
		allInitUsage, "",
		createTextUsage, createFileUsage, createAnyUsage, "",
		listTextUsage, listFilesUsage, listAnyUsage, "",
		getTextUsage, getFilesUsage, getAnyUsage, "",
		deleteTextUsage, deleteFilesUsage, deleteAnyUsage, "",
		listBackendsUsage, "",
		setDefaultBackendUsage, "",
		getkeyUsage, setkeyUsage,
	}
	allUsage = strings.Join(allUsages, "\n")
)
