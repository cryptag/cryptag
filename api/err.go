package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/elimisteve/cryptag/types"
)

const contentTypeJSON = "application/json; charset=utf-8"

func WriteJSONStatus(w http.ResponseWriter, obj interface{}, statusCode int) {
	writeJSON(w, obj, statusCode)
}

func WriteJSON(w http.ResponseWriter, obj interface{}) {
	writeJSON(w, obj, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, obj interface{}, statusCode int) {
	b, err := json.Marshal(obj)
	if err != nil {
		if types.Debug {
			log.Printf("Error marshaling `%#v`: %v\n", obj, err)
		}
		WriteError(w, "Error preparing response")
		return
	}

	if types.Debug {
		log.Printf("Writing JSON: `%s`\n", b)
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
	w.Write(b)
}

func WriteErrorStatus(w http.ResponseWriter, errStr string, statusCode int) {
	writeError(w, errStr, statusCode)
}

func WriteError(w http.ResponseWriter, errStr string) {
	writeError(w, errStr, http.StatusInternalServerError)
}

func writeError(w http.ResponseWriter, errStr string, status int) {
	if types.Debug {
		log.Printf("Returning HTTP %d w/error: %q\n", status, errStr)
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":%q}`, errStr)
}
