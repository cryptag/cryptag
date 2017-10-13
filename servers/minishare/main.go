// Steve Phillips / elimisteve
// 2017.01.02

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	minilock "github.com/cathalgarvey/go-minilock"
	"github.com/cathalgarvey/go-minilock/taber"
	"github.com/coreos/etcd/clientv3"
	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/api"
	gorillacontext "github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	uuid "github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
)

const (
	MINILOCK_ID_KEY      = "minilock_id"
	MINILOCK_KEYPAIR_KEY = "minilock_keypair"
)

var (
	Debug = false

	// TODO: Consider using a persistent key, which helps auth the
	// responses from this server
	randomServerKey *taber.Keys

	Expire1Day = int32((24 * time.Hour).Seconds())

	ErrAuthTokenNotFound = errors.New("Auth token not found")
	ErrSharesNotFound    = errors.New("Shares not found for that user")
)

func init() {
	if os.Getenv("DEBUG") == "1" {
		// Setting globlar var
		Debug = true
	}

	k, err := taber.RandomKey()
	if err != nil {
		log.Fatalf("Error generating random server key: %v\n", err)
	}

	// Setting globlar var
	randomServerKey = k
}

func main() {
	// etcd client
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Error creating etcd client: %v\n", err)
	}
	defer cli.Close()

	if Debug {
		log.Println("Connected to etcd")
	}

	// memcache client
	mc := memcache.New("localhost:11211")

	r := mux.NewRouter()

	// Client specifies claimed miniLock ID, gets back
	// miniLock-encrypted UUID as a challenge, then includes UUID in
	// auth header in subsequent requests
	r.HandleFunc("/login", Login(cli)).Methods("GET")

	// r.HandleFunc("/shares", auther(cli, GetShares(mc))).Methods("GET")
	r.HandleFunc("/shares/once", auther(cli, GetShares(mc))).Methods("GET")

	// TODO: Consider optionally requiring auth to share
	// r.HandleFunc("/shares", CreateShare(mc)).Methods("POST")
	r.HandleFunc("/shares/once", CreateShare(mc)).Methods("POST")

	// Evict this specific UUID from middleware (but not others for
	// this user, which may be used by other clients/apps/devices from
	// the same user)
	r.HandleFunc("/logout", auther(cli, Logout(cli))).Methods("POST")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./build"))).Methods("GET")

	http.Handle("/", r)

	listenAddr := getListenAddr()
	middleware := getMiddleware()

	log.Printf("Listening on %v\n", listenAddr)
	// TODO: Use server that times out
	log.Fatal(http.ListenAndServe(listenAddr, middleware.Then(r)))
}

func getListenAddr() string {
	listenAddr := "localhost:8000"
	if port := os.Getenv("PORT"); port != "" {
		listenAddr = "localhost:" + port
	}
	return listenAddr
}

func getMiddleware() alice.Chain {
	debugLogger := func(h http.Handler) http.Handler {
		if !Debug {
			return h
		}
		return handlers.LoggingHandler(os.Stderr, h)
	}

	return alice.New(debugLogger)
}

func auther(cli *clientv3.Client, h http.Handler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		authToken := parseAuthTokenFromHeader(req)
		if authToken == "" {
			writeErrorStatus(w, "Auth token missing",
				http.StatusUnauthorized, nil)
			return
		}

		mID, err := lookupMinilockIDByAuthToken(cli, authToken)
		if err != nil {
			status := http.StatusInternalServerError
			if err == ErrAuthTokenNotFound {
				status = http.StatusUnauthorized
			}
			writeErrorStatus(w, "Error authorizing you", status, err)
			return
		}

		if Debug {
			log.Printf("`%s` just authed; auth token: `%s`\n", mID, authToken)
		}

		// TODO: Update auth token's TTL/lease to be 1 hour from
		// _now_, not just 1 hour since when they first logged in

		keypair, err := taber.FromID(mID)
		if err != nil {
			writeError(w, "Your miniLock ID is invalid?...", err)
			return
		}

		gorillacontext.Set(req, MINILOCK_ID_KEY, mID)
		gorillacontext.Set(req, MINILOCK_KEYPAIR_KEY, keypair)

		h.ServeHTTP(w, req)
	}
}

func parseAuthTokenFromHeader(req *http.Request) string {
	bearerAndToken := req.Header.Get("Authorization")
	token := strings.TrimLeft(bearerAndToken, "Bearer ")
	return token
}

func lookupMinilockIDByAuthToken(cli *clientv3.Client, authToken string) (string, error) {
	timeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Look up miniLock ID by auth token
	resp, err := cli.Get(ctx, "token:"+authToken)
	if err != nil {
		return "", err
	}

	// Return miniLock ID
	for _, ev := range resp.Kvs {
		return string(ev.Value), nil
	}

	return "", ErrAuthTokenNotFound
}

func Login(cli *clientv3.Client) func(w http.ResponseWriter, req *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mID, keypair, err := parseMinilockID(req)
		if err != nil {
			writeErrorStatus(w, "Error: invalid miniLock ID",
				http.StatusBadRequest, err)
			return
		}

		if Debug {
			log.Printf("Login: `%s` is trying to log in\n", mID)
		}

		newUUID, err := uuid.NewV4()
		if err != nil {
			writeError(w, "Error generating now auth token; sorry!", err)
			return
		}

		authToken := newUUID.String()

		err = saveAuthToken(cli, mID, authToken)
		if err != nil {
			writeError(w, "Error saving new auth token; sorry!", err)
			return
		}

		filename := "auth_token-" + cryptag.NowStr()
		contents := []byte(authToken)
		sender := randomServerKey
		recipient := keypair

		encAuthToken, err := minilock.EncryptFileContents(filename, contents,
			sender, recipient)
		if err != nil {
			writeError(w, "Error encrypting auth token to you; sorry!", err)
			return
		}

		w.Write(encAuthToken)
	})
}

func parseMinilockID(req *http.Request) (string, *taber.Keys, error) {
	mID := req.Header.Get("X-Minilock-Id")

	// Validate miniLock ID by trying to generate public key from it
	keypair, err := taber.FromID(mID)
	if err != nil {
		return "", nil, fmt.Errorf("Error validating miniLock ID: %v", err)
	}

	return mID, keypair, nil
}

func saveAuthToken(cli *clientv3.Client, mID, authToken string) error {
	timeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Save authToken -> mID
	_, err := cli.Put(ctx, "token:"+authToken, mID)
	if err != nil {
		return err
	}

	return nil
}

func Logout(cli *clientv3.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Empty
		defer req.Body.Close()

		authToken := parseAuthTokenFromHeader(req)

		timeout := 3 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Delete authToken -> mID
		_, err := cli.Delete(ctx, "token:"+authToken)
		if err != nil {
			writeError(w, "Error logging you out", err)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func GetShares(mc *memcache.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Guaranteed by auther middleware
		mID := gorillacontext.Get(req, MINILOCK_ID_KEY).(string)

		shares, err := GetSharesByMinilockID(mc, mID)
		if err != nil {
			if err == ErrSharesNotFound {
				writeErrorStatus(w, err.Error(), http.StatusNotFound, err)
				return
			}
			writeError(w, "Error fetching shares", err)
			return
		}

		err = api.WriteJSON(w, shares)
		if err != nil {
			log.Printf("Error writing shares: %v\n", err)
			return
		}

		if !strings.Contains(req.URL.Path, "/once") {
			return
		}

		// User got shares and they should only be served/downloaded
		// once, so delete the ones we just served.

		// Avoiding a race condition by just deleting first N rather
		// than all of them, since a new one could have been shared
		// while we were doing the above and we wouldn't want to
		// delete that (unserved) one, too.
		err = DeleteFirstNSharesByMinilockID(mc, mID, len(shares))
		if err != nil {
			log.Printf("Error deleting first %d shares intended for %v: %v\n",
				len(shares), mID, err)
		}
	})
}

func CreateShare(mc *memcache.Client) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		// This is fine since we're storing body in memory anyway
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			writeError(w, "Error reading POST data", err)
			return
		}
		defer req.Body.Close()

		recipientIDs := req.Header["X-Minilock-Recipient-Ids"]
		if len(recipientIDs) == 0 {
			writeErrorStatus(w, "X-Minilock-Recipient-Ids header empty;"+
				" you cannot share with no one in particular",
				http.StatusBadRequest, nil)
			return
		}

		err = CreateShareForRecipients(mc, recipientIDs, body)
		if err != nil {
			writeError(w, "Error creating share", err)
			return
		}

		if Debug {
			log.Printf("New share successfully stored: `%s`",
				strings.Join(recipientIDs, ", "))
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func CreateShareForRecipients(mc *memcache.Client, recipientIDs []string, newshareb []byte) error {
	for _, rID := range recipientIDs {
		shares, err := GetSharesByMinilockID(mc, rID)
		if err != nil && err != ErrSharesNotFound {
			return err
		}

		shares = append(shares, newshareb)

		err = SetSharesByMinilockID(mc, rID, shares)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetSharesByMinilockID(mc *memcache.Client, mID string) ([][]byte, error) {
	if Debug {
		log.Printf("GetSharesByMinilockID for user `%s`\n", mID)
	}

	item, err := mc.Get("recipient_id:" + mID)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return nil, ErrSharesNotFound
		}
		return nil, err
	}

	var shares [][]byte

	if Debug {
		log.Printf("Unmarshaling share data of length %d\n", len(item.Value))
	}

	err = json.Unmarshal(item.Value, &shares)
	if err != nil {
		return shares, err
	}

	if len(shares) == 0 {
		return nil, ErrSharesNotFound
	}

	return shares, nil
}

func DeleteFirstNSharesByMinilockID(mc *memcache.Client, mID string, nShares int) error {
	shares, err := GetSharesByMinilockID(mc, mID)
	if err != nil {
		return err
	}

	if nShares > len(shares) {
		return fmt.Errorf("Cannot delete first %d of a total of %d existing"+
			" shares", nShares, len(shares))
	}

	remaining := shares[nShares:]

	return SetSharesByMinilockID(mc, mID, remaining)
}

func SetSharesByMinilockID(mc *memcache.Client, mID string, shares [][]byte) error {
	sharesb, err := json.Marshal(shares)
	if err != nil {
		return err
	}

	newItem := &memcache.Item{
		Key:        "recipient_id:" + mID,
		Value:      sharesb,
		Expiration: Expire1Day,
	}

	return mc.Set(newItem)
}
