package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cryptag/cryptag/types"
)

const contentTypeJSON = "application/json; charset=utf-8"

func WriteJSONBStatus(w http.ResponseWriter, jsonB []byte, statusCode int) error {
	if types.Debug {
		log.Printf("Writing JSON: `%s`\n", jsonB)
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
	_, err := w.Write(jsonB)
	if err != nil {
		log.Printf("Error writing response! %v\n", err)
	}
	return err
}

func WriteJSONB(w http.ResponseWriter, jsonB []byte) error {
	return WriteJSONBStatus(w, jsonB, http.StatusOK)
}

func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	return WriteJSONStatus(w, obj, http.StatusOK)
}

func WriteJSONStatus(w http.ResponseWriter, obj interface{}, statusCode int) error {
	b, err := json.Marshal(obj)
	if err != nil {
		if types.Debug {
			log.Printf("Error marshaling `%#v`: %v\n", obj, err)
		}
		WriteError(w, "Error preparing response")
		return err
	}

	return WriteJSONBStatus(w, b, statusCode)
}

func WriteError(w http.ResponseWriter, errStr string) error {
	return WriteErrorStatus(w, errStr, http.StatusInternalServerError)
}

func WriteErrorStatus(w http.ResponseWriter, errStr string, status int) error {
	if types.Debug {
		log.Printf("Returning HTTP %d w/error: %q\n", status, errStr)
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_, err := fmt.Fprintf(w, `{"error":%q}`, errStr)
	return err
}
