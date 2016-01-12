// Steve Phillips / elimisteve
// 2016.01.11

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		if len(os.Args) < 5 {
			log.Printf("At least 4 command line arguments must be included\n")
			log.Fatalf(createUsage)
		}

		//
		// Parse date
		//
		var when string

		date := strings.Split(os.Args[2], "/")
		switch length := len(date); length {
		case 0, 1:
			log.Fatalf("Date must be one of these forms: %s", validDateFormats)
		case 2:
			if len(date[0]) == 4 && len(date[1]) == 2 { // yyyy/mm
				when = date[0] + date[1] + "01"
			} else if len(date[0]) == 2 && len(date[1]) == 2 { // mm/dd
				thisYear := fmt.Sprintf("%v", time.Now().Year())
				when = thisYear + date[0] + date[1]
			} else {
				log.Fatalf("Invalid 2-part date `%v`\n", os.Args[2])
			}
		case 3:
			if len(date[0]) == 4 && len(date[1]) == 2 && len(date[2]) == 2 { // yyyy/mm/dd
				when = date[0] + date[1] + date[2]
			} else {
				log.Fatalf("Invalid 3-part date `%v`\n", os.Args[2])
			}
		default:
			log.Fatalf("Invalid date `%v`\n", os.Args[2])
		}

		todo := os.Args[3]
		tags := append(os.Args[4:], "when:"+when, "app:ccal", "type:todo",
			"type:text")

		if types.Debug {
			log.Printf("Creating row with data `%s` and tags `%#v`\n", todo, tags)
		}

		newRow, err := types.NewRow([]byte(todo), tags)
		if err != nil {
			log.Fatalf("Error creating new row: %v\n", err)
		}

		row, err := db.SaveRow(newRow)
		if err != nil {
			log.Fatalf("Error saving new row: %v\n", err)
		}
		fmt.Println(formatReminder(row))

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
		plaintags := append(os.Args[1:], "type:todo")
		rows, err := db.RowsFromPlainTags(plaintags)
		if err != nil {
			log.Fatal(err)
		}

		if len(rows) == 0 {
			log.Fatal(types.ErrRowsNotFound)
		}

		for _, r := range rows {
			fmt.Println(formatReminder(r))
		}
	}
}

func formatReminder(r *types.Row) string {
	return fmt.Sprintf(`%v  "%s"    %v`, types.RowTagWithPrefix(r, "when:"),
		r.Decrypted(), strings.Join(r.PlainTags(), "  "))
}

var (
	usage       = "Usage: " + filepath.Base(os.Args[0]) + " [create <date> <reminder> | delete] tag1 [tag2 ...]"
	createUsage = "Usage: " + filepath.Base(os.Args[0]) + " create <date> <reminder> tag1 [tag2 ...]"
	deleteUsage = "Usage: " + filepath.Base(os.Args[0]) + " delete tag1 [tag2 ...]"

	validDateFormats = "mm/dd, yyyy/mm, yyyy/mm/dd"
)
