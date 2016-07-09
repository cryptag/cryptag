// Steve Phillips / elimisteve
// 2016.06.23

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/api"
	"github.com/elimisteve/cryptag/api/trusted"
	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/cli"
	"github.com/elimisteve/cryptag/keyutil"
	"github.com/elimisteve/cryptag/types"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var backendName = "sandstorm-webserver"

func init() {
	if bn := os.Getenv("BACKEND"); bn != "" {
		backendName = bn
	}
}

func main() {
	var db backend.Backend

	db, err := backend.LoadWebserverBackend("", backendName)
	if err != nil {
		// TODO: Generically parse all Backend Configs
		log.Printf("Error from LoadWebserverBackend: %v", err)

		db, err = backend.LoadOrCreateFileSystem(
			os.Getenv("BACKEND_PATH"),
			os.Getenv("BACKEND"),
		)
		if err != nil {
			log.Fatalf("Error from LoadOrCreateFileSystem: %s", err)
		}
		log.Println("...but a FileSystem Backend loaded successfully")
	}

	if bk, ok := db.(cryptag.CanUseTor); ok && cryptag.UseTor {
		if err = bk.UseTor(); err != nil {
			log.Fatalf("Error trying to use Tor: %v\n", err)
		}
	}

	// Fetch and maintain up-to-date list of TagPairs
	pairs := NewTagPairStore()
	go func() {
		if err := pairs.Update(db); err != nil {
			log.Printf("Initial TagPair fetching failed! %v\n", err)
		}
	}()

	jsonNoError := map[string]string{"error": ""}

	Init := func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			api.WriteError(w, err.Error())
			return
		}
		defer req.Body.Close()

		m := map[string]string{}
		err = json.Unmarshal(body, &m)
		if err != nil {
			api.WriteErrorStatus(w, `Error parsing POST of the form`+
				` {"webkey": "..."}: `+err.Error(), http.StatusBadRequest)
			return
		}

		if err = cli.InitSandstorm(backendName, m["webkey"]); err != nil {
			api.WriteError(w, err.Error())
			return
		}

		api.WriteJSONStatus(w, jsonNoError, http.StatusCreated)
	}

	CreateRow := func(w http.ResponseWriter, req *http.Request) {
		// TODO: Do streaming reads
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			api.WriteError(w, err.Error())
			return
		}
		defer req.Body.Close()

		var trow trusted.Row
		err = json.Unmarshal(body, &trow)
		if err != nil {
			api.WriteErrorStatus(w, `Error parsing POST of the form`+
				` {"unencrypted": "(base64-encoded string)", "plaintags":`+
				` ["tag1", "tag2"]}`+err.Error(), http.StatusBadRequest)
			return
		}

		row, err := backend.CreateRow(db, pairs.Get(db), trow.Unencrypted, trow.PlainTags)
		if err != nil {
			api.WriteError(w, err.Error())
			return
		}

		go pairs.AsyncUpdate(db)

		// Return Row with null data, populated tags
		newTrow := trusted.Row{PlainTags: row.PlainTags()}
		trowB, _ := json.Marshal(&newTrow)

		api.WriteJSONBStatus(w, trowB, http.StatusCreated)
	}

	CreateFileRow := func(w http.ResponseWriter, req *http.Request) {
		// TODO: Do streaming reads
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			api.WriteError(w, err.Error())
			return
		}
		defer req.Body.Close()

		var trow trusted.FileRow
		err = json.Unmarshal(body, &trow)
		if err != nil {
			api.WriteErrorStatus(w, `Error parsing POST of the form`+
				` {"file_path": "/full/path/to/file", "plaintags":`+
				` ["tag1", "tag2"]}`+err.Error(), http.StatusBadRequest)
			return
		}

		row, err := backend.CreateFileRow(db, pairs.Get(db), trow.FilePath, trow.PlainTags)
		if err != nil {
			api.WriteError(w, err.Error())
			return
		}

		go pairs.AsyncUpdate(db)

		// Return Row with null data, populated tags
		newTrow := trusted.Row{PlainTags: row.PlainTags()}
		trowB, _ := json.Marshal(&newTrow)

		api.WriteJSONBStatus(w, trowB, http.StatusCreated)
	}

	GetKey := func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, `{"key":[%s]}`, keyutil.Format(db.Key()))
	}

	SetKey := func(w http.ResponseWriter, req *http.Request) {
		keyB, err := ioutil.ReadAll(req.Body)
		if err != nil {
			api.WriteError(w, err.Error())
			return
		}
		defer req.Body.Close()

		newKey, err := keyutil.Parse(string(keyB))
		if err != nil {
			api.WriteErrorStatus(w, "Error parsing key: "+err.Error(),
				http.StatusBadRequest)
			return
		}

		if err = backend.UpdateKey(db, newKey); err != nil {
			api.WriteError(w, "Error updating key: "+err.Error())
			return
		}

		api.WriteJSONStatus(w, jsonNoError, http.StatusCreated)
	}

	ListRows := func(w http.ResponseWriter, req *http.Request) {
		plaintags, handledReq := parsePlaintags(w, req)
		if handledReq {
			return
		}

		rows, err := backend.ListRowsFromPlainTags(db, pairs.Get(db), plaintags)
		if err != nil {
			api.WriteError(w, err.Error())
			return
		}

		rowsB, err := json.Marshal(trusted.FromRows(rows))
		if err != nil {
			api.WriteError(w, err.Error())
			return
		}

		api.WriteJSONB(w, rowsB)
	}

	GetRows := func(w http.ResponseWriter, req *http.Request) {
		plaintags, handledReq := parsePlaintags(w, req)
		if handledReq {
			return
		}

		rows, err := backend.RowsFromPlainTags(db, pairs.Get(db), plaintags)
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "found") {
				api.WriteErrorStatus(w, errStr, http.StatusNotFound)
				return
			}
			api.WriteError(w, errStr)
			return
		}

		rowsB, err := json.Marshal(trusted.FromRows(rows))
		if err != nil {
			api.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
			return
		}

		api.WriteJSONB(w, rowsB)
	}

	GetTags := func(w http.ResponseWriter, req *http.Request) {
		newPairs, err := db.AllTagPairs(pairs.Get(db))
		if err != nil {
			api.WriteError(w, "Error fetching tag pairs: "+err.Error())
			return
		}

		pairs.Set(db, newPairs)

		pairsB, err := json.Marshal(trusted.FromTagPairs(newPairs))
		if err != nil {
			api.WriteError(w, "Error marshaling tag pairs: "+err.Error())
			return
		}

		api.WriteJSONB(w, pairsB)
	}

	DeleteRows := func(w http.ResponseWriter, req *http.Request) {
		plaintags, handledReq := parsePlaintags(w, req)
		if handledReq {
			return
		}

		if len(plaintags) == 0 {
			api.WriteErrorStatus(w, "No plaintags included in query",
				http.StatusBadRequest)
			return
		}

		if err = backend.DeleteRows(db, pairs.Get(db), plaintags); err != nil {
			api.WriteError(w, "Error deleting rows: "+err.Error())
			return
		}

		api.WriteJSONStatus(w, jsonNoError, http.StatusCreated)
	}

	// Mount handlers to router

	r := mux.NewRouter()

	r.HandleFunc("/trusted/init", Init).Methods("POST")

	r.HandleFunc("/trusted/rows/get", GetRows).Methods("POST")
	r.HandleFunc("/trusted/rows", CreateRow).Methods("POST")
	r.HandleFunc("/trusted/rows/file", CreateFileRow).Methods("POST")
	r.HandleFunc("/trusted/rows/list", ListRows).Methods("POST")
	r.HandleFunc("/trusted/rows/delete", DeleteRows).Methods("POST")

	r.HandleFunc("/trusted/tags", GetTags).Methods("GET")

	r.HandleFunc("/trusted/key", GetKey).Methods("GET")
	r.HandleFunc("/trusted/key", SetKey).Methods("POST")

	http.Handle("/", r)

	logger := func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(os.Stderr, h)
	}
	middleware := alice.New(logIncomingReq, logger)

	listenAddr := "localhost:7878"
	log.Printf("Listening on %v\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, middleware.Then(r)))
}

func logIncomingReq(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("INCOMING: %v %v\n", req.Method, req.URL.Path)
		h.ServeHTTP(w, req)
	})
}

func parsePlaintags(w http.ResponseWriter, req *http.Request) (plaintags []string, handledReq bool) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.WriteError(w, err.Error())
		return nil, true
	}
	defer req.Body.Close()

	err = json.Unmarshal(body, &plaintags)
	if err != nil {
		errStr := fmt.Sprintf("Error parsing POSTed JSON array of tags `%s`: %s",
			body, err)
		api.WriteErrorStatus(w, errStr, http.StatusBadRequest)
		return nil, true
	}

	return plaintags, false
}

func NewTagPairStore() *TagPairStore {
	store := &TagPairStore{pairs: map[string]types.TagPairs{}}
	// go store.loop()
	return store
}

type TagPairStore struct {
	mu sync.RWMutex

	// map[backendName]pairs
	pairs map[string]types.TagPairs
}

func (store *TagPairStore) Update(bk backend.Backend) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	oldPairs := store.pairs[bk.Name()]

	newPairs, err := bk.AllTagPairs(oldPairs)
	if err != nil {
		return fmt.Errorf("Error updating %s's TagPairs: %v", bk.Name(), err)
	}

	store.pairs[bk.Name()] = newPairs
	return nil
}

func (store *TagPairStore) AsyncUpdate(bk backend.Backend) {
	if err := store.Update(bk); err != nil {
		log.Println(err)
	}
}

func (store *TagPairStore) Get(bk backend.Backend) types.TagPairs {
	store.mu.RLock()
	defer store.mu.RUnlock()

	return store.pairs[bk.Name()]
}

func (store *TagPairStore) Set(bk backend.Backend, newPairs types.TagPairs) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.pairs[bk.Name()] = newPairs
}
