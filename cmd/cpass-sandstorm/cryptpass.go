// Steve Phillips / elimisteve
// 2015.02.24

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/backend"
	"github.com/cryptag/cryptag/cli/color"
	"github.com/elimisteve/clipboard"
)

var backendName = "sandstorm-webserver"

func init() {
	if bn := os.Getenv("BACKEND"); bn != "" {
		backendName = bn
	}
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln(usage)
	}

	var db *backend.WebserverBackend

	if os.Args[1] != "init" {
		var err error
		db, err = backend.LoadWebserverBackend("", backendName)
		if err != nil {
			log.Printf("%v\n", err)
			log.Fatal(usage)
		}
	}

	if cryptag.UseTor {
		err := db.UseTor()
		if err != nil {
			log.Fatalf("Error trying to use Tor: %v\n", err)
		}
	}

	switch os.Args[1] {
	case "init":
		if len(os.Args) < 3 {
			log.Fatal(initUsage)
		}
		if err := createBackendConfig(os.Args[2]); err != nil {
			log.Fatal(err)
		}

	case "create":
		if len(os.Args) < 4 {
			log.Println("At least 3 command line arguments must be included")
			log.Fatal(usage)
		}

		data := os.Args[2]
		tags := append(os.Args[3:], "app:cryptpass", "type:text")

		row, err := backend.CreateRow(db, nil, []byte(data), tags)
		if err != nil {
			log.Fatalf("Error creating then saving new row: %v", err)
		}

		color.Println(color.TextRow(row))

	case "getkey":
		fmt.Println(fmtKey(db.Key()))

	case "setkey":
		if len(os.Args) < 3 {
			log.Println("At least 2 command line arguments must be included")
			log.Fatal(usage)
		}

		keyStr := strings.Join(os.Args[2:], ",")

		newKey, err := parseKey(keyStr)
		if err != nil {
			log.Fatalf("Error from parseKey: %v\n", err)
		}

		cfg, err := db.ToConfig()
		if err != nil {
			log.Fatal(err)
		}

		cfg.Key = newKey

		if err := cfg.Update(cryptag.BackendPath); err != nil {
			log.Fatalf("Error updating config: %v", err)
		}

	default: // Search
		// Empty clipboard
		clipboard.WriteAll(nil)

		plaintags := append(os.Args[1:], "type:text")

		rows, err := backend.RowsFromPlainTags(db, nil, plaintags)
		if err != nil {
			log.Fatal(err)
		}

		// Add first row's contents to clipboard
		dec := rows[0].Decrypted()
		if err = clipboard.WriteAll(dec); err != nil {
			log.Printf("Error writing first result to clipboard: %v\n", err)
		} else {
			log.Printf("Added first result `%s` to clipboard\n", dec)
		}

		color.Println(color.TextRows(rows))
	}
}

var usage = "Usage: " + filepath.Base(os.Args[0]) + " [create <yourpassword>] tag1 [tag2 ...]"
var initUsage = "Usage: " + filepath.Base(os.Args[0]) + " init <sandstorm_webkey>"

func createBackendConfig(key string) error {
	info := strings.SplitN(key, "#", 2)
	if len(info) < 2 {
		return fmt.Errorf(
			"Error parsing `%v` as Sandstorm key generated from Sandstorm's UI",
			key)
	}

	serverBaseURL, authToken := info[0], info[1]

	db, err := backend.NewWebserverBackend(nil, backendName, serverBaseURL, authToken)
	if err != nil {
		return fmt.Errorf("NewWebserverBackend error: %v\n", err)
	}

	cfg, err := db.ToConfig()
	if err != nil {
		return fmt.Errorf("Error getting backend config: %v\n", err)
	}

	err = cfg.Save(cryptag.BackendPath)
	if err != nil && err != backend.ErrConfigExists {
		return fmt.Errorf("Error saving backend config to disk: %v\n", err)
	}

	return nil
}

var keyRegex = regexp.MustCompile(`(\d+)`)

func parseKey(cliDigits string) (*[32]byte, error) {
	// Pluck out all digit sequences, convert to numbers
	nums := keyRegex.FindAllString(cliDigits, -1)
	if len(nums) != 32 {
		return nil, fmt.Errorf("Key must include 32 numbers, not %d", len(nums))
	}

	var newKey [32]byte

	for i := 0; i < 32; i++ {
		n, err := strconv.ParseUint(nums[i], 10, 8)
		if err != nil {
			return nil, fmt.Errorf("Number #%d '%v' was invalid: %v\n", i+1,
				nums[i])
		}
		newKey[i] = byte(n)
	}

	return &newKey, nil
}

func fmtKey(key *[32]byte) string {
	if key == nil {
		return "<nil>"
	}
	k := *key

	kStr := fmt.Sprintf("%d", k[0])

	for i := 0; i < len(k)-1; i++ {
		kStr += fmt.Sprintf(",%d", k[i+1])
	}
	return kStr
}
