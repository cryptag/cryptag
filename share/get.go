// Steve Phillips / elimisteve
// 2017.01.14

package share

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	minilock "github.com/cathalgarvey/go-minilock"
	"github.com/cathalgarvey/go-minilock/taber"
	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/backend"
	"github.com/cryptag/cryptag/types"
)

var (
	ErrInvalidInviteURL = errors.New("share: invalid invite URL")
)

// Share represents a(n originally miniLock-encrypted) "share" --
// something that has been shared via a share server.
type Share struct {
	SenderID string
	Filename string
	Contents []byte
}

func GetConfigsByInviteURL(url string) ([]*backend.Config, error) {
	return toConfigs(GetSharesByInviteURL(url))
}

func GetSharesByInviteURL(inviteURL string) ([]*Share, error) {
	serverBaseURL, passphrase, err := ParseInviteURL(inviteURL)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(serverBaseURL, ".onion") {
		cryptag.UseTor = true
	}

	authToken := ""
	return GetShares(serverBaseURL, authToken, EmailFromPassphrase(passphrase),
		passphrase)
}

func ParseInviteURL(url string) (serverBaseURL, passphrase string, err error) {
	strs := strings.Split(url, "#")
	if len(strs) != 2 {
		return "", "", ErrInvalidInviteURL
	}
	// TODO: Consider doing more validity checks
	serverBaseURL = strings.TrimSuffix(strs[0], "/")
	passphrase = strs[1]
	return serverBaseURL, passphrase, nil
}

// GetShares fetches all shares for the user with the given email and
// passphrase.
func GetShares(serverBaseURL string, authToken, recipEmail, recipPassphrase string) ([]*Share, error) {
	keypair, err := taber.FromEmailAndPassphrase(recipEmail, recipPassphrase)
	if err != nil {
		return nil, err
	}

	serverBaseURL = strings.TrimSuffix(serverBaseURL, "/")

	return GetSharesByKeypair(serverBaseURL, authToken, keypair)
}

func GetConfigsByKeypair(serverBaseURL, authToken string, keypair *taber.Keys) ([]*backend.Config, error) {
	return toConfigs(GetSharesByKeypair(serverBaseURL, authToken, keypair))
}

// GetShares fetches all shares for the user with the given keypair.
func GetSharesByKeypair(serverBaseURL, authToken string, keypair *taber.Keys) ([]*Share, error) {
	resp, err := Get(serverBaseURL, "/shares/once", authToken, keypair)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: Stream the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Wanted HTTP %d, got %d; body: %s",
			http.StatusOK, resp.StatusCode, body)
	}

	// Slice of miniLock-encrypted files
	var allsharesb [][]byte

	err = json.Unmarshal(body, &allsharesb)
	if err != nil {
		return nil, err
	}

	shares := make([]*Share, 0, len(allsharesb))
	var errs []string

	for _, shareb := range allsharesb {
		senderID, filename, contents, err := minilock.DecryptFileContents(
			shareb, keypair)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		share := &Share{
			SenderID: senderID,
			Filename: filename,
			Contents: contents,
		}
		shares = append(shares, share)
	}

	if len(errs) != 0 {
		err = fmt.Errorf("share: %d error(s) decrypting shares: %s",
			len(errs), strings.Join(errs, "; "))
	}

	return shares, err
}

// Get does an authenticated GET request to serverBaseURL+path. If
// authToken is empty or stale, this function will log into the server
// to get a new auth token, and _then_ execute the GET.
func Get(serverBaseURL, path, authToken string, keypair *taber.Keys) (*http.Response, error) {
	if authToken == "" {
		token, err := Login(serverBaseURL, keypair)
		if err != nil {
			return nil, err
		}
		if types.Debug {
			log.Printf("Get: logged in; decrypted auth token: `%v`\n", token)
		}
		authToken = token
	}

	req, err := http.NewRequest("GET", serverBaseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+authToken)

	resp, err := getClient().Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		if types.Debug {
			log.Println("Unauthorized GET; logging in again to get new auth token")
		}

		newAuthToken, err := Login(serverBaseURL, keypair)
		if err != nil {
			return nil, err
		}

		req.Header["Authorization"] = []string{"Bearer " + newAuthToken}

		return getErrChecks(getClient().Do(req))
	}

	return getErrChecks(resp, err)
}

func getErrChecks(resp *http.Response, err error) (*http.Response, error) {
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusNotFound {
			m := map[string]string{}
			err = json.Unmarshal(body, &m)
			if err != nil {
				return nil, fmt.Errorf("Error parsing 404: %v", err)
			}
			return nil, errors.New(m["error"])
		}

		return nil, fmt.Errorf("Got HTTP %v, wanted %v; response: %s",
			resp.StatusCode, http.StatusOK, body)
	}

	return resp, nil
}

func Login(serverBaseURL string, keypair *taber.Keys) (authToken string, err error) {
	if types.Debug {
		log.Printf("Logging into server `%s`\n", serverBaseURL)
	}
	req, err := http.NewRequest("GET", serverBaseURL+"/login", nil)
	if err != nil {
		return "", err
	}

	mID, err := keypair.EncodeID()
	if err != nil {
		return "", err
	}

	req.Header.Add("X-Minilock-Id", mID)

	resp, err := getClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	encAuthToken, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	_, _, contents, err := minilock.DecryptFileContents(encAuthToken, keypair)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

//
// Helpers
//

func ToConfigs(shares []*Share) ([]*backend.Config, error) {
	configs := make([]*backend.Config, 0, len(shares))

	var errs []string

	for _, share := range shares {
		cfg, err := ToConfig(share)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		configs = append(configs, cfg)
	}

	var err error
	if len(errs) != 0 {
		err = fmt.Errorf("share: %d error(s) converting shares to configs: %s",
			len(errs), strings.Join(errs, "; "))
	}

	return configs, err
}

func toConfigs(shares []*Share, err error) ([]*backend.Config, error) {
	if err != nil {
		return nil, err
	}
	return ToConfigs(shares)
}

func ToConfig(share *Share) (*backend.Config, error) {
	cfg := &backend.Config{}
	err := json.Unmarshal(share.Contents, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
