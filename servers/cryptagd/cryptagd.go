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
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/api"
	"github.com/cryptag/cryptag/api/trusted"
	"github.com/cryptag/cryptag/backend"
	"github.com/cryptag/cryptag/keyutil"
	"github.com/cryptag/cryptag/rowutil"
	"github.com/cryptag/cryptag/types"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var backendName = ""

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
		log.Println("No backends found")
		bk, err := backend.LoadOrCreateDefaultFileSystemBackend(
			os.Getenv("BACKEND_PATH"),
			os.Getenv("BACKEND"),
		)
		if err != nil {
			log.Fatalf("No Backends successfully read! Failed to create "+
				"one: %v", err)
		}
		log.Printf("...but this one was loaded or created: %v\n", bk.Name())
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
	// pairs.AsyncUpdateAll(backends)

	jsonNoError := map[string]string{"error": ""}

	Init := func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			api.WriteError(w, err.Error())
			return
		}
		defer req.Body.Close()

		var cfg backend.Config
		err = json.Unmarshal(body, &cfg)
		if err != nil {
			api.WriteErrorStatus(w, `Error parsing POST of the form`+
				` {"Name": "...", "Type": "..."}: `+err.Error(),
				http.StatusBadRequest)
			return
		}

		bk, err := backend.CreateFromConfig("", &cfg)
		if err != nil {
			api.WriteError(w, "Error creating new Backend Config: "+err.Error())
			return
		}

		if err := bkStore.Add(bk); err != nil {
			statusCode := http.StatusInternalServerError
			if err == backend.ErrBackendExists {
				statusCode = http.StatusConflict
			}
			api.WriteErrorStatus(w, "Error adding new Backend after creation `"+
				bk.Name()+"`: "+err.Error(), statusCode)
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
				` ["tag1", "tag2"]} -- `+err.Error(), http.StatusBadRequest)
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

	UpdateRow := func(w http.ResponseWriter, req *http.Request) {
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

		var trowu trusted.RowUpdate
		err = json.Unmarshal(body, &trowu)
		if err != nil {
			api.WriteErrorStatus(w, `Error parsing POST of the form`+
				` {"unencrypted": "(base64-encoded string)", "old_version_id_tag":`+
				` "id:..."}`+err.Error(), http.StatusBadRequest)
			return
		}

		row, err := backend.UpdateRow(db, pairs.Get(db), trowu.OldVersionID,
			trowu.Unencrypted)
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

	UpdateFileRow := func(w http.ResponseWriter, req *http.Request) {
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

		var trowu trusted.FileRowUpdate
		err = json.Unmarshal(body, &trowu)
		if err != nil {
			api.WriteErrorStatus(w, `Error parsing POST of the form`+
				` {"file_path": "/full/path/to/file", "old_version_id_tag":`+
				` "id:..."}`+err.Error(), http.StatusBadRequest)
			return
		}

		// OldVersionID can be the ID of _any_ previous version of this file

		row, err := backend.UpdateFileRow(db, pairs.Get(db), trowu.OldVersionID,
			trowu.FilePath)
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

		trows := getTrustedRowsByPath(req.URL.Path, rows)

		rowsB, err := json.Marshal(trows)
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

		trows := getTrustedRowsByPath(req.URL.Path, rows)

		rowsB, err := json.Marshal(trows)
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

	GetBackends := func(w http.ResponseWriter, req *http.Request) {
		bkHeader := req.Header.Get("X-Backend")
		if bkHeader == "" {
			bkHeader = "*"
		}

		// TODO: Once there are .json.minilock files there, parse
		// bkPattern+"*.json.minilock" then decrypt
		configs, err := backend.ReadConfigs("", bkHeader)
		if err != nil {
			if len(configs) == 0 {
				api.WriteError(w, "Error reading Backend Configs: "+err.Error())
				return
			}

			log.Printf("Error reading some Backend configs: %v\n", err)

			// FALL THROUGH
		}

		tconfigs := trusted.FromConfigs(configs)

		api.WriteJSON(w, tconfigs)
	}

	GetBackendNames := func(w http.ResponseWriter, req *http.Request) {
		bkPattern := req.Header.Get("X-Backend")
		if bkPattern == "" {
			bkPattern = "*"
		}

		bkFile := filepath.Join(cryptag.BackendPath, bkPattern+".json")

		bkFilenames, err := filepath.Glob(bkFile)
		if err != nil {
			errStr := fmt.Sprintf("Error globbing Configs with pattern `%s`: %v",
				bkPattern, err)
			api.WriteError(w, errStr)
			return
		}

		bkNames := make([]string, 0, len(bkFilenames))

		for _, fname := range bkFilenames {
			bkNames = append(bkNames, backend.ConfigNameFromPath(fname))
		}

		api.WriteJSON(w, bkNames)
	}

	// Mount handlers to router

	r := mux.NewRouter()

	r.HandleFunc("/trusted/init", Init).Methods("POST")

	r.HandleFunc("/trusted/rows/get", GetRows).Methods("POST")
	r.HandleFunc("/trusted/rows/get/versioned", GetRows).Methods("POST")
	r.HandleFunc("/trusted/rows/get/versioned/latest", GetRows).Methods("POST")
	r.HandleFunc("/trusted/rows", CreateRow).Methods("POST")
	r.HandleFunc("/trusted/rows/file", CreateFileRow).Methods("POST")
	r.HandleFunc("/trusted/rows/list", ListRows).Methods("POST")
	r.HandleFunc("/trusted/rows/list/versioned", ListRows).Methods("POST")
	r.HandleFunc("/trusted/rows/list/versioned/latest", ListRows).Methods("POST")
	r.HandleFunc("/trusted/rows/delete", DeleteRows).Methods("POST")

	r.HandleFunc("/trusted/rows", UpdateRow).Methods("PUT")
	r.HandleFunc("/trusted/rows/file", UpdateFileRow).Methods("PUT")

	r.HandleFunc("/trusted/tags", GetTags).Methods("GET")

	r.HandleFunc("/trusted/key", GetKey).Methods("GET")
	r.HandleFunc("/trusted/key", SetKey).Methods("POST")

	r.HandleFunc("/trusted/backends", GetBackends).Methods("GET")
	r.HandleFunc("/trusted/backends/names", GetBackendNames).Methods("GET")

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
	bk, err := bkStore.Get(bkHeader, backendName, "default")
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

	// If there's only 1 Backend, use it
	if len(store.bks) == 1 {
		for _, bk := range store.bks {
			return bk, nil
		}
	}

	return nil, fmt.Errorf("No Backend with any of these names: %s",
		strings.Join(bkNames, ", "))
}

func (store *BackendStore) Add(bk backend.Backend) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.bks[bk.Name()]; exists {
		return backend.ErrBackendExists
	}

	store.bks[bk.Name()] = bk
	return nil
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

// TODO: For efficiency, don't fetch every version of every Row when
// we only care about the most recent version of each
func getTrustedRowsByPath(urlPath string, rows types.Rows) (trows interface{}) {
	if !strings.Contains(urlPath, "/versioned") {
		return trusted.FromRows(rows)
	}

	// Segment rows into versioned rows

	vrows := rowutil.ToVersionedRows(rows, rowutil.ByTagPrefix("created:", false))
	tvrows := trusted.FromRows2D(vrows)

	if !strings.Contains(urlPath, "/latest") {
		return tvrows
	}

	latests := make(trusted.Rows, 0, len(tvrows))
	for _, rows := range tvrows {
		latests = append(latests, rows[len(rows)-1])
	}

	return latests
}
