// Steven Phillips / elimisteve
// 2016.06.16

package main

import (
	"log"
	"net/http"

	"github.com/cryptag/cryptag/api"
)

func writeError(w http.ResponseWriter, errStr string, secretErr error) {
	log.Printf("Returning HTTP %d w/error: %q;\n  real error: %s\n",
		http.StatusInternalServerError, errStr, secretErr)

	api.WriteError(w, errStr)
}

func writeErrorStatus(w http.ResponseWriter, errStr string, status int, secretErr error) {
	log.Printf("Returning HTTP %d w/error: %q; real error: %s\n", status,
		errStr, secretErr)

	api.WriteErrorStatus(w, errStr, status)
}
