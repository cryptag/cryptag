// Steve Phillips / elimisteve
// 2013.08.12

package fun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	httpclient "github.com/mreiferson/go-httpclient"
)

var (
	HttpPostTimeout = 60 * time.Second

	UserAgentStrings = []string{
		"Mozilla/5.0 (Windows NT 6.1; rv:10.0) Gecko/20100101 Firefox/10.0",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:22.0) Gecko/20100101 Firefox/22.0",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/28.0.1500.72 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:22.0) Gecko/20100101 Firefox/22.0",
	}
)

func Fetch(url string, timeout time.Duration) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating GET request: %v", err)
	}

	// Change User Agent string (default: "Go http package")
	request.Header.Set("User-Agent", RandomUserAgent())

	// Create an HTTP client that times out after timeout seconds
	transport := &httpclient.Transport{RequestTimeout: timeout}
	defer transport.Close()

	client := &http.Client{Transport: transport}

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func FetchInto(url string, timeout time.Duration, structure interface{}) error {
	body, err := Fetch(url, timeout)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, structure)
}

func Post(url string, data []byte) (*http.Response, error) {
	request, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Error creating POST request: %v", err)
	}

	transport := &httpclient.Transport{RequestTimeout: HttpPostTimeout}
	defer transport.Close()

	client := &http.Client{Transport: transport}

	return client.Do(request)
}

func RandomUserAgent() string {
	ndx := rand.Intn(len(UserAgentStrings))
	return UserAgentStrings[ndx]
}
