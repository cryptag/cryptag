// Steve Phillips / elimisteve
// 2012.06.09
// 2014.01.05

package fun

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func Scrape(url string) ([]byte, error) {
	// Download contents of `url`
	req, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error GET'ing %s: %s\n", url, err)
	}
	defer req.Body.Close()

	// Save contents of page to `html` variable
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading from req.Body: %s\n", err)
	}
	return data, nil
}

func SimpleHTTPServer(handler http.Handler, listenAddr string) *http.Server {
	return &http.Server{
		Addr:           listenAddr,
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}
