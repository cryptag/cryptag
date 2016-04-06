// Steve Phillips / elimisteve
// 2016.01.11

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
		log.Fatalln(usage)
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 5 {
			log.Printf("At least 4 command line arguments must be included\n")
			log.Fatalf(createUsage)
		}

		when, err := parseDate(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}

		todo := os.Args[3]
		tags := append(os.Args[4:], "when:"+when, "app:cremind",
			"type:calendarevent", "type:text")

		if types.Debug {
			log.Printf("Creating row with data `%s` and tags `%#v`\n", todo, tags)
		}

		row, err := types.NewRow([]byte(todo), tags)
		if err != nil {
			log.Fatalf("Error creating new row: %v\n", err)
		}

		err = db.SaveRow(row)
		if err != nil {
			log.Fatalf("Error saving new row: %v\n", err)
		}
		fmt.Println(fmtReminder(row))

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
		tags := os.Args[1:]

		if tags[0] == "today" {
			tags[0] = "when:" + fmtDate(time.Now())
		}

		plaintags := append(tags, "type:calendarevent")

		// TODO: Cache tags locally
		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		rows, err := backend.RowsFromPlainTags(db, plaintags, pairs)
		if err != nil {
			log.Fatal(err)
		}

		// Sort and print

		events := make([]string, 0, len(rows))
		for _, r := range rows {
			events = append(events, fmtReminder(r))
		}
		sort.Strings(events)
		color.Println(strings.Join(events, "\n\n"))
	}
}

var (
	timeLayout = "20060102"
)

func fmtReminder(r *types.Row) string {
	whenStr := types.RowTagWithPrefix(r, "when:")
	when, err := time.Parse(timeLayout, whenStr)
	if err != nil {
		log.Printf("Error parsing timestamp `%v` as format `%v`: %v\n", whenStr,
			timeLayout, err)
	}

	// Indicate whether this event is planned for today
	today := ""
	if isToday(when) {
		today = "*"
	}

	dayOfWeek := when.Weekday().String()[:3] // E.g., "Fri"
	return fmt.Sprintf(`%s %s "%s"    %s`, color.BlackOnCyan(whenStr),
		color.BlackOnCyan(dayOfWeek+today), r.Decrypted(),
		color.Tags(r.PlainTags()))
}

func isToday(day time.Time) bool {
	y, m, d := day.Date()
	y2, m2, d2 := time.Now().Date()
	return y == y2 && m == m2 && d == d2
}

func fmtDate(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%d%02d%02d", y, m, d)
}

func parseDate(dateOrig string) (string, error) {
	if dateOrig == "today" {
		_, month, day := time.Now().Date()
		dateOrig = fmt.Sprintf("%d/%d", month, day)
	}

	date := strings.Split(dateOrig, "/")

	switch length := len(date); length {
	case 0, 1:
		return "", fmt.Errorf("Date must be in one of these formats: %s",
			validDateFormats)
	case 2:
		if year := date[0]; len(year) == 4 { // yyyy/m(m)
			month, err := validMonth(date[1])
			if err != nil {
				return "", err
			}
			day := "01"

			return year + month + day, nil
		}

		// m(m)/d(d)
		month, err := validMonth(date[0])
		if err != nil {
			return "", err
		}
		day, err := validDay(date[1])
		if err != nil {
			return "", err
		}
		year := fmt.Sprintf("%v", time.Now().Year())

		return year + month + day, nil
	case 3: // yyyy/m(m)/d(d)
		year := date[0]
		if len(year) != 4 {
			return "", fmt.Errorf("Invalid year `%v`", year)
		}
		month, err := validMonth(date[1])
		if err != nil {
			return "", err
		}
		day, err := validDay(date[2])
		if err != nil {
			return "", err
		}

		return year + month + day, nil
	}

	return "", fmt.Errorf("Invalid date `%v`\n", os.Args[2])
}

func validMonth(month string) (string, error) {
	if len(month) == 2 {
		return month, nil
	}
	if len(month) == 1 {
		return "0" + month, nil
	}
	return "", fmt.Errorf("Invalid month `%v`", month)
}

func validDay(day string) (string, error) {
	if len(day) == 2 {
		return day, nil
	}
	if len(day) == 1 {
		return "0" + day, nil
	}
	return "", fmt.Errorf("Invalid day `%v`", day)
}

var (
	usage       = "Usage: " + filepath.Base(os.Args[0]) + " [create <date> <reminder> | delete] tag1 [tag2 ...]"
	createUsage = "Usage: " + filepath.Base(os.Args[0]) + " create <date> <reminder> tag1 [tag2 ...]"
	deleteUsage = "Usage: " + filepath.Base(os.Args[0]) + " delete tag1 [tag2 ...]"

	validDateFormats = "'today', m(m)/d(d), yyyy/m(m), yyyy/m(m)/d(d)"
)
