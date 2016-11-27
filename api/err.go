package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cryptag/cryptag/types"
)

const contentTypeJSON = "application/json; charset=utf-8"

func WriteJSONBStatus(w http.ResponseWriter, jsonB []byte, statusCode int) {
	if types.Debug {
		log.Printf("Writing JSON: `%s`\n", jsonB)
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
	w.Write(jsonB)
}

func WriteJSONB(w http.ResponseWriter, jsonB []byte) {
	WriteJSONBStatus(w, jsonB, http.StatusOK)
}

func WriteJSON(w http.ResponseWriter, obj interface{}) {
	WriteJSONStatus(w, obj, http.StatusOK)
}

func WriteJSONStatus(w http.ResponseWriter, obj interface{}, statusCode int) {
	b, err := json.Marshal(obj)
	if err != nil {
		if types.Debug {
			log.Printf("Error marshaling `%#v`: %v\n", obj, err)
		}
		WriteError(w, "Error preparing response")
		return
	}

	WriteJSONBStatus(w, b, statusCode)
}

func WriteError(w http.ResponseWriter, errStr string) {
	WriteErrorStatus(w, errStr, http.StatusInternalServerError)
}

func WriteErrorStatus(w http.ResponseWriter, errStr string, status int) {
	if types.Debug {
		log.Printf("Returning HTTP %d w/error: %q\n", status, errStr)
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":%q}`, errStr)
}
