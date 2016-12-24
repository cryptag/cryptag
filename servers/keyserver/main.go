// Steven Phillips / elimisteve
// 2016.06.15

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"github.com/cryptag/cryptag/api"
	"github.com/gorilla/mux"
)

var Debug = false

func init() {
	if os.Getenv("DEBUG") == "1" {
		Debug = true
	}
}

func main() {
	db := mustInitDB()
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/minilock/{email}", GetMinilockID(db)).Methods("GET")
	r.HandleFunc("/minilock", PostMinilockID(db)).Methods("POST")

	http.Handle("/", r)

	listenAddr := ":8000"
	if port := os.Getenv("PORT"); port != "" {
		listenAddr = ":" + port
	}

	log.Printf("Listening on %v\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, r))
}

func GetMinilockID(db *bolt.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		email := mux.Vars(req)["email"]
		mID, err := GetMinilockIDByEmail(db, email)
		if err != nil {
			log.Printf("Error from GetMinilockIDByEmail(%q): %v\n", email, err)
			if err == ErrMinilockIDNotFound {
				writeErrorStatus(w, err.Error(), http.StatusNotFound, err)
				return
			}
			writeError(w, "Error fetching miniLock ID", err)
			return
		}
		ident := &Identity{Email: email, MinilockID: mID}
		api.WriteJSON(w, ident)
		return
	}
}

func PostMinilockID(db *bolt.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			writeError(w, "Error reading POST data", err)
			return
		}
		defer req.Body.Close()

		var ident Identity
		_ = json.Unmarshal(body, &ident)
		if !ident.Valid() {
			writeErrorStatus(w, "Invalid JSON; populate 'email' and 'minilock_id'",
				http.StatusBadRequest, nil)
			return
		}

		err = CreateMinilockIDByEmail(db, ident.Email, ident.MinilockID)
		if err != nil {
			log.Printf("Error from CreateMinilockIDByEmail(%q, %q): %v\n",
				ident.Email, ident.MinilockID, err)
			if err == ErrMinilockIDExists {
				writeErrorStatus(w, err.Error(), http.StatusConflict, err)
				return
			}
			writeError(w, "Error saving new miniLock ID", err)
			return
		}

		if Debug {
			log.Printf("New MinilockID successfully stored: %s -> %s\n",
				ident.Email, ident.MinilockID)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

type Identity struct {
	Email      string `json:"email,omitempty"`
	MinilockID string `json:"minilock_id,omitempty"`
}

func (id *Identity) Valid() bool {
	if id == nil {
		return false
	}
	// TODO: More sophisticated checks. 'Email' technically need not
	// be an email, though.
	return id.Email != "" && id.MinilockID != ""
}
