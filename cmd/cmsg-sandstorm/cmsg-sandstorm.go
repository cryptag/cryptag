// Steven Phillips / elimisteve
// 2016.07.03

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

	case "createroom":
		// 0:cryptmessage 1:createroom 2:<roomname> 3:[<tag1> ...]
		if len(os.Args) < 3 {
			cli.ArgFatal(createroomUsage)
		}

		roomName := os.Args[2]
		plaintags := append(os.Args[3:], "app:cryptmessage", "type:chatroom",
			"roomname:"+roomName)

		row, err := backend.CreateRow(db, nil, nil, plaintags)
		if err != nil {
			log.Fatalf("Error creating then saving new row: %v", err)
		}

		color.Println(color.TextRow(row))

	case "send":
		// 0:cryptmessage 1:send 2:<from> 3:<room name> 4:<msg> 5:[<tag1> ...]
		if len(os.Args) < 5 {
			cli.ArgFatal(sendUsage)
		}

		from := os.Args[2]
		roomName := os.Args[3]
		msg := os.Args[4]

		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		roomPlaintags := cryptag.PlainTags{"type:chatroom", "roomname:" + roomName}

		rows, err := backend.ListRowsFromPlainTags(db, pairs, roomPlaintags)
		if err != nil {
			log.Fatal(err)
		}

		switch {
		case len(rows) == 0:
			log.Fatalf("No chatroom named '%v'", roomName)
		case len(rows) > 1:
			color.Printf("%d chatrooms exist with that name: %v", len(rows),
				color.TextRows(rows))
		}

		// len(rows) == 1

		roomIDTag := rowutil.TagWithPrefix(rows[0], "id:")

		// TODO: Add miniLock-encrypted sha256(msg) to this struct to
		// basically cryptographically sign messages. Clients should
		// decrypt then verify that the message is from who it says
		// it's from. Users who join later will not be among the
		// recipients and therefore won't be able to verify.
		message := Message{Msg: msg}

		msgPlaintags := append(os.Args[5:], "app:cryptmessage",
			"type:chatmessage", "parentrow:"+roomIDTag, "from:"+from)

		row, err := backend.CreateJSONRow(db, pairs, &message, msgPlaintags)
		if err != nil {
			log.Fatalf("Error creating then saving new row: %v", err)
		}

		color.Println(color.TextRow(row))

	case "viewroom":
		// 0:cryptmessage 1:viewroom 2:<roomname> 3:[<tag1> ...]
		roomName := os.Args[2]
		plaintags := append(os.Args[3:], "type:chatroom", "roomname:"+roomName)

		pairs, err := db.AllTagPairs()
		if err != nil {
			log.Fatal(err)
		}

		rows, err := backend.ListRowsFromPlainTags(db, pairs, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		if len(rows) != 1 {
			log.Fatalf("Wanted 1 room, got %d instead\n", len(rows))
		}

		plaintags = []string{
			"type:chatmessage",
			"parentrow:" + rowutil.TagWithPrefix(rows[0], "id:"),
		}

		rows, err = backend.RowsFromPlainTags(db, pairs, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		rows.Sort(rowutil.ByTagPrefix("created:", true))
		rowStrs := rowutil.MapToStrings(fmtMsg, rows)
		color.Println(strings.Join(rowStrs, "\n\n"))

	case "listrooms":
		plaintags := append(os.Args[2:], "type:chatroom")

		rows, err := backend.ListRowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		rows.Sort(rowutil.ByTagPrefix("created:", true))
		rowStrs := rowutil.MapToStrings(fmtRoom, rows)
		color.Println(strings.Join(rowStrs, "\n\n"))

	case "deleteroom", "deletemsg":
		if len(os.Args) < 3 {
			if os.Args[1] == "deleteroom" {
				cli.ArgFatal(deleteroomUsage)
			}
			cli.ArgFatal(deletemsgUsage)
		}

		typeTag := "type:chatmessage"
		if os.Args[1] == "deletemsg" {
			typeTag = "type:chatroom"
			// First arg after subcommand is room name
			os.Args[2] = "roomname:" + os.Args[2]
		}

		plainTags := append(os.Args[2:], typeTag)

		err := backend.DeleteRows(db, nil, plainTags)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Row(s) successfully deleted\n")

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

	case "getmsgs": // Search
		plaintags := append(os.Args[2:], "type:chatmessage")

		rows, err := backend.RowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		rows.Sort(rowutil.ByTagPrefix("created:", true))
		rowStrs := rowutil.MapToStrings(fmtMsg, rows)
		color.Println(strings.Join(rowStrs, "\n\n"))

	default:
		log.Printf("Subcommand `%s` not valid\n", os.Args[1])
		cli.Fatal(allUsage)
	}
}

func fmtMsg(r *types.Row) string {
	var msg Message
	err := json.Unmarshal(r.Decrypted(), &msg)
	if err != nil {
		return fmt.Sprintf("(Error unmarshaling Message row: %v)", err)
	}

	from := rowutil.TagWithPrefixStripped(r, "from:")

	// May or may not contain "to:" tag
	to := rowutil.TagWithPrefixStripped(r, "to:")
	if to != "" {
		to = color.BlackOnCyan("To") + ": " + to + "\n"
	}

	return fmt.Sprintf(`%s%s: %s
%s`,
		to,
		color.BlackOnCyan(from), msg.Msg,
		color.Tags(r.PlainTags()))
}

func fmtRoom(r *types.Row) string {
	roomName := rowutil.TagWithPrefixStripped(r, "roomname:")
	id := rowutil.TagWithPrefix(r, "id:")

	return fmt.Sprintf(`%s: %s
%s:   %s
%s`,
		color.BlackOnCyan("Room Name"), roomName,
		color.BlackOnCyan("Room ID"), id,
		color.Tags(r.PlainTags()))
}

type Message struct {
	Msg string `json:"msg"`
}

var (
	prefix = "Usage: " + filepath.Base(os.Args[0]) + " "

	initUsage       = prefix + "init <sandstorm_webkey>"
	createroomUsage = prefix + "createroom <name> [<tag1> ...]"
	sendUsage       = prefix + "send <from> <room name> <msg> [<tag1> ...]"
	viewroomUsage   = prefix + "viewroom <room name> [<tag1> ...]"
	listroomsUsage  = prefix + "listrooms [<tag1> ...]"
	getmsgsUsage    = prefix + "getmsgs [roomname:... from:... <tag3> ...]"
	deleteroomUsage = prefix + "deleteroom <room name> [<tag1> ...]"
	deletemsgUsage  = prefix + "deletemsg <tag1> [<tag2> ...]"
	getkeyUsage     = prefix + "getkey"
	setkeyUsage     = prefix + "setkey <32-number crypto key>"

	allUsages = []string{initUsage, createroomUsage, sendUsage, viewroomUsage,
		listroomsUsage, getmsgsUsage, deleteroomUsage, deletemsgUsage,
		getkeyUsage, setkeyUsage}

	allUsage = strings.Join(allUsages, "\n")
)
