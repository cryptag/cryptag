// Steven Phillips / elimisteve
// 2016.06.16

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func writeJSON(w http.ResponseWriter, obj interface{}) {
	b, err := json.Marshal(obj)
	if err != nil {
		log.Printf("Error marshaling `%#v`: %v\n", obj, err)
		writeError(w, "Error preparing response", http.StatusInternalServerError)
		return
	}

	log.Printf("Writing JSON: `%s`\n", b)

	w.Header().Set("Context-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func writeError(w http.ResponseWriter, errStr string, status int) {
	log.Printf("Returning HTTP %d w/error: %q\n", status, errStr)

	w.Header().Set("Context-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	io.WriteString(w, fmt.Sprintf(`{"error": %q}`, errStr))
}
