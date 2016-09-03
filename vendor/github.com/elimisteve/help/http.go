// Steve Phillips / elimisteve
// 2014.10.15

package help

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, structure interface{}) {
	jsonData, err := json.Marshal(structure)
	if err != nil {
		WriteError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonData)
}

func ReadInto(rBody io.ReadCloser, structure interface{}) error {
	body, err := ReadAndClose(rBody)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, structure)
}

func ReadAndClose(rBody io.ReadCloser) ([]byte, error) {
	body, err := ioutil.ReadAll(rBody)
	defer rBody.Close()
	if err != nil {
		return nil, err
	}

	return body, nil
}
