// Steve Phillips / elimisteve
// 2017.01.13

package share

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	minilock "github.com/cathalgarvey/go-minilock"
	"github.com/cryptag/cryptag/backend"
)

const (
	DefaultServerURL         = "https://minishare.cryptag.org"
	DefaultServerTorURL      = "http://ptga662wjtg2cie3.onion"
	DefaultWordsInPassphrase = 12
)

// CreateEphemeral uses passphrase to use miniLock to encrypt
// JSON-marshaled cfg and store it at the share server at
// serverBaseURL, returning the location at which the Backend Config
// (aka invite) can be retrieved.
//
// If serverBaseURL is empty, DefaultServerURL is used.
func CreateEphemeral(serverBaseURL string, cfg *backend.Config) (shareURL string, err error) {
	cfgb, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}

	passphrase, err := RandomPassphrase(DefaultWordsInPassphrase)
	if err != nil {
		return "", err
	}

	// TODO: Consider ensuring that cfg.Name is as desired
	filename := cfg.Name + ".json"

	keypair, err := NewKeyPair(passphrase)
	if err != nil {
		return "", err
	}

	// This is how ephemeral Shares work
	sender := keypair
	recipient := keypair

	fileb, err := minilock.EncryptFileContents(filename, cfgb, sender, recipient)
	if err != nil {
		return "", err
	}

	recipientID, err := recipient.EncodeID()
	if err != nil {
		return "", err
	}

	cl := NewClient(serverBaseURL)

	err = Post(cl, "/shares/once", bytes.NewReader(fileb),
		recipientHeaders([]string{recipientID}))
	if err != nil {
		return "", err
	}

	return BuildShareURL(cl.ServerBaseURL, passphrase), nil
}

func recipientHeaders(recips []string) http.Header {
	return http.Header{"X-Minilock-Recipient-Ids": recips}
}

// Posts reads filebr and POSTs it to the Share server at
// cl.ServerBaseURL + urlPath.
func Post(cl *Client, urlPath string, filebr io.Reader, headers http.Header) error {
	req, err := http.NewRequest("POST", cl.ServerBaseURL+urlPath, filebr)
	if err != nil {
		return err
	}
	req.Header = headers

	resp, err := cl.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Wanted HTTP %d, got %d; body: %s",
			http.StatusCreated, resp.StatusCode, body)
	}

	return nil
}

// BuildShareURL builds and returns the final URL at which the Share
// server at serverBaseURL is hosting Shares for the user whose
// keypair can be generated with passphrase.
func BuildShareURL(serverBaseURL string, passphrase string) string {
	serverBaseURL = strings.TrimRight(serverBaseURL, "/")
	return serverBaseURL + "/#" + passphrase
}
