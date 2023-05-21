// Steven Phillips / elimisteve
// 2016.04.17

package tor

import (
	"bytes"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

type urlTextTest struct {
	url, text string
}

var tests = []urlTextTest{
	// Each URL and the string that site should contain
	{"http://duskgytldkxiuqc6.onion/", "Tor"},
	{"http://checkip.dyndns.org/", "IP Address"},
}

func TestDoRequest(t *testing.T) {
	if os.Getenv("RUN_TOR_TESTS") != "1" {
		t.Skip()
	}

	wg := sync.WaitGroup{}
	wg.Add(len(tests))

	for _, test := range tests {
		go func(test urlTextTest) {
			defer wg.Done()

			t.Logf("URL: %s\n", test.url)

			resp, err := DoRequest(nil, "GET", test.url, nil)
			if err != nil {
				t.Errorf("Error creating new request: %v", err)
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Error reading response body: %v", err)
				return
			}

			if !bytes.Contains(body, []byte(test.text)) {
				t.Errorf("Tor page doesn't contain `%s` like it should. Body: `%s`",
					test.text, body)
				return
			}

			t.Logf("Text `%s` found at URL `%s`\n", test.text, test.url)
		}(test)
	}

	wg.Wait()

}

func TestDoRequestUnparseableProxy(t *testing.T) {
	ProxyURL = ""

	_, err := NewClient()
	if err == nil {
		t.Fatal("Didn't get error creating client, should have")
	}
}

func TestDoRequestNonexistentProxy(t *testing.T) {
	ProxyURL = "socks5://127.0.0.1:1234"

	_, err := DoRequest(nil, "GET", tests[0].url, nil)
	if err == nil {
		t.Fatal("DoRequest didn't fail with non-existent proxy, should have")
	}
}
