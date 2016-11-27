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
	"regexp"
	"strings"
	"sync"

	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/api"
	"github.com/cryptag/cryptag/api/trusted"
	"github.com/cryptag/cryptag/backend"
	"github.com/cryptag/cryptag/cli"
	"github.com/cryptag/cryptag/keyutil"
	"github.com/cryptag/cryptag/types"
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
	backends, err := backend.ReadBackends("", "*")
	if err != nil {
		log.Printf("Error reading Backends: %v\n", err)

		// Fall through
	}

	if len(backends) == 0 {
		bk, err := backend.LoadOrCreateFileSystem(
			os.Getenv("BACKEND_PATH"),
			os.Getenv("BACKEND"),
		)
		if err != nil {
			log.Fatalf("No Backends successfully read! Failed to create "+
				"one: %v", err)
		}
		log.Printf("...but a new one was created: %v\n", bk.Name())
		backends = []backend.Backend{bk}
	}

	for _, bk := range backends {
		if tbk, ok := bk.(cryptag.CanUseTor); ok && cryptag.UseTor {
			if err = tbk.UseTor(); err != nil {
				log.Fatalf("Error telling %s to use Tor: %v\n", bk.Name(), err)
			}
			if types.Debug {
				log.Printf("Backend %s successfully told to use Tor\n", bk.Name())
			}
		}
	}

	// map[bk.Name()]bk
	bkStore := NewBackendStore(backends)

	// Fetch and maintain up-to-date list of TagPairs
	pairs := NewTagPairStore()
	pairs.AsyncUpdateAll(backends)

	jsonNoError := map[string]string{"error": ""}

	Init := func(w http.ResponseWriter, req *http.Request) {
		bkName := req.Header.Get("X-Backend")

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

		// TODO: Create generic init
		if err = cli.InitSandstorm(bkName, m["webkey"]); err != nil {
			api.WriteError(w, err.Error())
			return
		}

		api.WriteJSONStatus(w, jsonNoError, http.StatusCreated)
	}

	CreateRow := func(w http.ResponseWriter, req *http.Request) {
		db, handledReq := getBackend(bkStore, w, req)
		if handledReq {
			return
		}

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
		db, handledReq := getBackend(bkStore, w, req)
		if handledReq {
			return
		}

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
		db, handledReq := getBackend(bkStore, w, req)
		if handledReq {
			return
		}

		fmt.Fprintf(w, `{"key":[%s]}`, keyutil.Format(db.Key()))
	}

	SetKey := func(w http.ResponseWriter, req *http.Request) {
		db, handledReq := getBackend(bkStore, w, req)
		if handledReq {
			return
		}

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
		db, handledReq := getBackend(bkStore, w, req)
		if handledReq {
			return
		}

		plaintags, handledReq := parsePlaintags(w, req)
		if handledReq {
			return
		}

		rows, err := fetchRowsFromPlainTags(backend.ListRowsFromPlainTags, db, pairs, plaintags)
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
			api.WriteError(w, err.Error())
			return
		}

		api.WriteJSONB(w, rowsB)
	}

	GetRows := func(w http.ResponseWriter, req *http.Request) {
		db, handledReq := getBackend(bkStore, w, req)
		if handledReq {
			return
		}

		plaintags, handledReq := parsePlaintags(w, req)
		if handledReq {
			return
		}

		rows, err := fetchRowsFromPlainTags(backend.RowsFromPlainTags, db, pairs, plaintags)
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
		db, handledReq := getBackend(bkStore, w, req)
		if handledReq {
			return
		}

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
		db, handledReq := getBackend(bkStore, w, req)
		if handledReq {
			return
		}

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "7878"
	}

	listenAddr := "localhost:" + port

	log.Printf("Listening on %v\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, middleware.Then(r)))
}

func logIncomingReq(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("INCOMING: %v %v\n", req.Method, req.URL.Path)
		h.ServeHTTP(w, req)
	})
}

func getBackend(bkStore *BackendStore, w http.ResponseWriter, req *http.Request) (bk backend.Backend, handledReq bool) {
	bkHeader := req.Header.Get("X-Backend")
	if types.Debug {
		log.Printf("X-Backend parsed as: `%v`\n", bkHeader)
	}
	bk, err := bkStore.Get(bkHeader, backendName, "sandstorm-webserver", "default")
	if err != nil {
		api.WriteError(w, "Error fetching Backend: "+err.Error())
		return nil, true
	}

	return bk, false
}

func parsePlaintags(w http.ResponseWriter, req *http.Request) (plaintags []string, handledReq bool) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.WriteError(w, err.Error())
		return nil, true
	}
	defer req.Body.Close()

	var creq Request

	err = json.Unmarshal(body, &creq)
	if err != nil {
		errStr := fmt.Sprintf(`Error parsing POSTed JSON object with`+
			` 'plaintags' key %s: %s`, body, err)
		api.WriteErrorStatus(w, errStr, http.StatusBadRequest)
		return nil, true
	}

	// TODO: Return error if len(req.PlainTags) == 0?

	return creq.PlainTags, false
}

type Request struct {
	PlainTags []string `json:"plaintags"`
}

//
// BackendStore
//

type BackendStore struct {
	mu sync.RWMutex

	// map[backendName]Backend
	bks map[string]backend.Backend
}

func NewBackendStore(bks []backend.Backend) *BackendStore {
	bkMap := map[string]backend.Backend{}

	for _, bk := range bks {
		bkMap[bk.Name()] = bk
	}

	store := &BackendStore{bks: bkMap}
	return store
}

func (store *BackendStore) Get(bkPrimary string, bkNames ...string) (backend.Backend, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if bkPrimary != "" {
		bk := store.bks[bkPrimary]
		if bk == nil {
			return nil, fmt.Errorf("Backend `%s` not found", bkPrimary)
		}
		return bk, nil
	}

	for _, name := range bkNames {
		bk := store.bks[name]
		if bk != nil {
			if types.Debug {
				log.Printf("BackendStore.Get: returning backend `%s`\n", bk.Name())
			}
			return bk, nil
		}
		if types.Debug {
			log.Printf("Backend `%s` not found\n", name)
		}
	}

	return nil, fmt.Errorf("No Backend with any of these names: %s",
		strings.Join(bkNames, ", "))
}

//
// TagPairStore
//

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
	store.mu.RLock()
	oldPairs := store.pairs[bk.Name()]
	store.mu.RUnlock()

	newPairs, err := bk.AllTagPairs(oldPairs)
	if err != nil {
		return fmt.Errorf("Error updating %s's TagPairs: %v", bk.Name(), err)
	}

	store.mu.Lock()
	store.pairs[bk.Name()] = newPairs
	store.mu.Unlock()

	return nil
}

func (store *TagPairStore) AsyncUpdate(bk backend.Backend) {
	if err := store.Update(bk); err != nil {
		log.Println(err)
	}
}

func (store *TagPairStore) AsyncUpdateAll(bks []backend.Backend) {
	for _, bk := range bks {
		go store.AsyncUpdate(bk)
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

//
// If Row-fetching error was caused by out-of-date TagPairs, fetch
// them and retry
//

func fetchRowsFromPlainTags(fetcher func(backend.Backend, types.TagPairs, cryptag.PlainTags) (types.Rows, error), bk backend.Backend, pairStore *TagPairStore, plaintags []string) (types.Rows, error) {
	rows, err := fetcher(bk, pairStore.Get(bk), plaintags)
	if err == nil {
		return rows, nil
	}

	if match, _ := regexp.MatchString("(?:Random|Plain)Tag `[a-z0-9]+?` not found", err.Error()); match {
		if err = pairStore.Update(bk); err != nil {
			return nil, fmt.Errorf("Error re-fetching TagPairs: %v", err)
		}
		return fetcher(bk, pairStore.Get(bk), plaintags)
	}

	return nil, err
}
