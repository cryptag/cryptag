// Steven Phillips / elimisteve
// 2016.06.04

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/cli"
	"github.com/elimisteve/cryptag/cli/color"
	"github.com/elimisteve/cryptag/keyutil"
	"github.com/elimisteve/cryptag/mobile/cryptask"
	"github.com/elimisteve/cryptag/rowutil"
	"github.com/elimisteve/cryptag/types"
)

var (
	backendName = "sandstorm-webserver"
)

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
			log.Fatalf("Error loading webserver config: %v", err)
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

	case "create":
		// 0:cryptask 1:create 2:<title> 3:<description> 4:[<assignee:NAME> <tag2> ...]
		if len(os.Args) < 4 {
			cli.ArgFatal(createUsage)
		}

		task := cryptask.Task{Title: os.Args[2], Description: os.Args[3]}
		plaintags := append(os.Args[4:], "app:cryptask", "type:task")

		row, err := backend.CreateJSONRow(db, nil, &task, plaintags)
		if err != nil {
			log.Fatalf("Error creating then saving new row: %v", err)
		}

		color.Println(fmtTask(row))

	case "delete":
		if len(os.Args) < 3 {
			cli.ArgFatal(deleteUsage)
		}

		plainTags := append(os.Args[2:], "app:cryptask", "type:task")

		err := backend.DeleteRows(db, nil, plainTags)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Task(s) successfully deleted\n")

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

	case "get": // Search
		plaintags := append(os.Args[2:], "app:cryptask", "type:task")

		rows, err := backend.RowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		rows.Sort(rowutil.ByTagPrefix("created:", true))

		// Format each row according to fmtTask, print result
		rowStrs := rowutil.MapToStrings(fmtTask, rows)
		color.Println(strings.Join(rowStrs, "\n\n\n"))

	default:
		log.Printf("Subcommand `%s` not valid\n", os.Args[1])
		cli.Fatal(allUsage)
	}
}

func fmtTask(r *types.Row) string {
	var task cryptask.Task
	err := json.Unmarshal(r.Decrypted(), &task)
	if err != nil {
		return fmt.Sprintf("(Error unmarshaling task row: %v)", err)
	}

	// Look for `assignee:...` tags. If they exist, show who the talk
	// is assigned to.

	assignees := rowutil.TagsWithPrefixStripped(r, "assignee:")
	var assignee string

	if len(assignees) > 0 {
		assigneesStr := strings.Join(assignees, ", ")
		assignee = "\n" + color.BlackOnCyan("Assignee(s)") + ": " + assigneesStr
	}

	return fmt.Sprintf(`%s:  %s
%s: %s%s
%s`,
		color.BlackOnCyan("Task Title"), task.Title,
		color.BlackOnCyan("Description"), task.Description, assignee,
		color.Tags(r.PlainTags()))
}

var (
	prefix = "Usage: " + filepath.Base(os.Args[0]) + " "

	initUsage   = prefix + "init <sandstorm_key>"
	createUsage = prefix + "create <title> <description> [assignee:NAME] [<tag2> ...]"
	getUsage    = prefix + "get [assignee:NAME] [<tag2> ...]"
	deleteUsage = prefix + "delete <tag1> [id:UUID] [<tag3> ...]"
	getkeyUsage = prefix + "getkey"
	setkeyUsage = prefix + "setkey <32-number crypto key>"

	allUsages = []string{initUsage, createUsage, getUsage,
		deleteUsage, getkeyUsage, setkeyUsage}

	allUsage = strings.Join(allUsages, "\n")
)
