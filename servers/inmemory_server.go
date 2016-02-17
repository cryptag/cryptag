// Steve Phillips / elimisteve
// 2015.02.24

package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/elimisteve/cryptag/types"
	"github.com/elimisteve/fun"
	"github.com/elimisteve/help"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// Rows
	router.HandleFunc("/rows", GetRows).Methods("GET")
	router.HandleFunc("/rows", PostRow).Methods("POST")

	// Tags
	router.HandleFunc("/tags", GetTags).Methods("GET")
	router.HandleFunc("/tags", PostTag).Methods("POST")

	http.Handle("/", router)

	server := fun.SimpleHTTPServer(router, ":7777")
	log.Printf("HTTP server trying to listen on %v...\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func GetRows(w http.ResponseWriter, req *http.Request) {
	_ = req.ParseForm()

	tags := req.Form["tags"]
	if len(tags) == 0 {
		log.Printf("All %d Rows retrieved:\n%s", len(allRows), allRows)
		help.WriteJSON(w, allRows)
		return
	}
	tags = strings.Split(tags[0], ",")

	if types.Debug {
		log.Printf("tags: `%+v`\n", tags)
	}

	rows := allRows.HaveAllRandomTags(tags)

	if types.Debug {
		log.Printf("%d/%d Rows retrieved:\n%s", len(rows), len(allRows), rows)
	}

	help.WriteJSON(w, rows)
}

func PostRow(w http.ResponseWriter, req *http.Request) {
	row := &types.Row{}
	if err := help.ReadInto(req.Body, row); err != nil {
		help.WriteError(w, "Error reading rows: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	allRows = append(allRows, row)
	if types.Debug {
		log.Printf("New row added: `%#v`\n", row)
	}

	help.WriteJSON(w, row)
}

func GetTags(w http.ResponseWriter, req *http.Request) {
	tags := req.Form["tags"]
	if len(tags) == 0 {
		if types.Debug {
			log.Printf("All %d TagPairs retrieved", len(allTagPairs))
		}
		help.WriteJSON(w, allTagPairs)
		return
	}
	tags = strings.Split(tags[0], ",")

	pairs, _ := allTagPairs.HaveAllRandomTags(tags)

	if types.Debug {
		log.Printf("%d TagPairs retrieved", len(pairs))
	}
	help.WriteJSON(w, pairs)
}

func PostTag(w http.ResponseWriter, req *http.Request) {
	pair := &types.TagPair{}
	if err := help.ReadInto(req.Body, pair); err != nil {
		help.WriteError(w, "Error reading tag pair: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	allTagPairs = append(allTagPairs, pair)
	if types.Debug {
		log.Printf("New TagPair added: `%#v`\n", pair)
	}

	help.WriteJSON(w, pair)
}

//
// Replace with Postgres or S3
//

var allTagPairs = types.TagPairs{}

var allRows = types.Rows{}
