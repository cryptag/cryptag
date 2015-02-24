// Steve Phillips / elimisteve
// 2015.02.24

package main

import (
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/222Labs/help"
	"github.com/elimisteve/cryptag/types"
	"github.com/elimisteve/fun"
	"github.com/gorilla/mux"
)

var secretRoot = ""

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	router := mux.NewRouter()

	// Rows
	router.HandleFunc(secretRoot, GetRows).Methods("GET")
	router.HandleFunc(secretRoot, PostRow).Methods("POST")

	// Tags
	router.HandleFunc(secretRoot+"/tags", GetTags).Methods("GET")
	router.HandleFunc(secretRoot+"/tags", PostTag).Methods("POST")

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

	log.Printf("tags == `%#v`\n", tags)
	rows := allRows.HaveAllRandomTags(tags)

	log.Printf("%d/%d Rows retrieved:\n%s", len(rows), len(allRows), rows)
	help.WriteJSON(w, rows)
}

func PostRow(w http.ResponseWriter, req *http.Request) {
	row := &types.Row{}
	if err := help.ReadInto(req.Body, row); err != nil {
		http.Error(w, "Error reading rows: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	allRows = append(allRows, row)
	log.Printf("New row added: `%#v`\n", row)

	help.WriteJSON(w, row)
}

func GetTags(w http.ResponseWriter, req *http.Request) {
	help.WriteJSON(w, allTagPairs)
}

func PostTag(w http.ResponseWriter, req *http.Request) {
	pair := &types.TagPair{}
	if err := help.ReadInto(req.Body, pair); err != nil {
		http.Error(w, "Error reading tag pair: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	allTagPairs = append(allTagPairs, pair)
	log.Printf("New TagPair added: `%#v`\n", pair)

	help.WriteJSON(w, pair)
}

//
// Replace with Postgres or S3
//

var allTagPairs = types.TagPairs{}

var allRows = types.Rows{}
