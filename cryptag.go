// Steve Phillips / elimisteve
// 2015.02.24

package main

import (
	"crypto/aes"
	"fmt"
	"log"
	"os"

	"github.com/elimisteve/cryptag/types"
)

var (
	SERVER_BASE_URL = ""
	SHARED_SECRET   = ""
)

func init() {
	// Set global `Block` for AES encryption/decryption
	block, err := aes.NewCipher([]byte(SHARED_SECRET))
	if err != nil {
		log.Fatalf("Error from aes.NewCipher: %v\n", err)
	}

	// TODO: Clean this up
	types.Block = block
	types.SERVER_BASE_URL = SERVER_BASE_URL
	types.Debug = true
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalf(usage)
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 4 {
			log.Printf("At least 3 command line arguments must be included\n")
			log.Fatalf(usage)
		}

		data := os.Args[2]
		tags := os.Args[3:]
		row := types.NewRow([]byte(data), tags)

		if err := row.Save(); err != nil {
			log.Fatalf("Error saving new row: %v\n", err)
		}
		fmt.Print(row.Format())

	default: // Search
		plaintags := os.Args[1:]
		rows, err := types.FetchByPlainTags(plaintags)
		if err != nil {
			log.Fatalf("Error from FetchByPlainTags: %v\n", err)
		}
		if err = rows.FirstToClipboard(); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing first result to clipboard: %v\n", err)
		}
		fmt.Print(rows.Format())
	}
}

var usage = "Usage: " + os.Args[0] + " data tag1 [tag2 ...]"
