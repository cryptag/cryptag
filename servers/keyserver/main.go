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
	"github.com/gorilla/mux"
)

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
				writeError(w, err.Error(), http.StatusNotFound)
				return
			}
			writeError(w, "Error fetching miniLock ID",
				http.StatusInternalServerError)
			return
		}
		ident := &Identity{Email: email, MinilockID: mID}
		writeJSON(w, ident)
		return
	}
}

func PostMinilockID(db *bolt.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("Error reading body: %v\n", err)
			writeError(w, "Unknown error", http.StatusInternalServerError)
			return
		}

		var ident Identity
		_ = json.Unmarshal(body, &ident)
		if !ident.Valid() {
			writeError(w, "Invalid JSON; populate 'email' and 'minilock_id'",
				http.StatusBadRequest)
			return
		}

		err = CreateMinilockIDByEmail(db, ident.Email, ident.MinilockID)
		if err != nil {
			log.Printf("Error from CreateMinilockIDByEmail(%q, %q): %v\n",
				ident.Email, ident.MinilockID, err)
			if err == ErrMinilockIDExists {
				writeError(w, err.Error(), http.StatusConflict)
				return
			}
			writeError(w, "Error saving new miniLock ID",
				http.StatusInternalServerError)
			return
		}

		log.Printf("New MinilockID successfully stored: %s -> %s\n",
			ident.Email, ident.MinilockID)

		w.WriteHeader(http.StatusCreated)
	}
}

type Identity struct {
	Email      string `json:"email",omitempty`
	MinilockID string `json:"minilock_id",omitempty`
}

func (id *Identity) Valid() bool {
	if id == nil {
		return false
	}
	// TODO: More sophisticated checks. 'Email' technically need not
	// be an email, though.
	return id.Email != "" && id.MinilockID != ""
}
