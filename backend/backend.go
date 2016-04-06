// Steve Phillips / elimisteve
// 2015.03.01

package backend

import (
	"errors"
	"fmt"
	"log"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/types"
	"github.com/elimisteve/fun"
)

var (
	RANDOM_TAG_ALPHABET = "abcdefghijklmnopqrstuvwxyz0123456789"
	RANDOM_TAG_LENGTH   = 9
)

type Backend interface {
	Key() *[32]byte

	AllTagPairs() (types.TagPairs, error)
	TagPairsFromRandomTags(randtags cryptag.RandomTags) (types.TagPairs, error)
	SaveTagPair(pair *types.TagPair) error

	ListRows(randtags cryptag.RandomTags) (types.Rows, error)
	RowsFromRandomTags(randtags cryptag.RandomTags) (types.Rows, error)
	SaveRow(row *types.Row) error
	DeleteRows(randtags cryptag.RandomTags) error
}

func CreateTagsFromPlain(backend Backend, plaintags []string) (allPairs types.TagPairs, newPairs types.TagPairs, err error) {
	// Fetch all tags
	allPairs, err = backend.AllTagPairs()
	if err != nil {
		return nil, nil, fmt.Errorf("Error from AllTagPairs: %v", err)
	}

	if types.Debug {
		log.Printf("Fetched all %d TagPairs from server\n", len(allPairs))
	}

	// Find out which members of plaintags don't have an existing,
	// corresponding TagPair

	existingPlain := allPairs.AllPlain()

	// Concurrent Tag creation ftw
	var chs []chan *types.TagPair

	// TODO: Put the following in a `CreateTags` function

	for _, plain := range plaintags {
		if !fun.SliceContains(existingPlain, plain) {
			// Preserve tag ordering despite concurrent creation
			ch := make(chan *types.TagPair)
			chs = append(chs, ch)

			go func(plain string, ch chan *types.TagPair) {
				pair, err := CreateTag(backend, plain)
				if err != nil {
					log.Printf("Error calling CreateTag(%q): %v\n", plain, err)
					ch <- nil
					return
				}
				if types.Debug {
					log.Printf("Created tag pair `%#v` (%p)\n", pair, pair)
				}
				ch <- pair
				return
			}(plain, ch)
		}
	}

	// Append successfully-created *TagPair values to `chs`
	//
	// TODO: Consider timing out in case CreateTag() never returns
	for i := 0; i < len(chs); i++ {
		if p := <-chs[i]; p != nil {
			newPairs = append(newPairs, p)
		}
	}

	// TODO(elimisteve): WTF?

	allPairs = append(allPairs, newPairs...)

	return allPairs, newPairs, nil
}

func NewTagPair(key *[32]byte, plaintag string) (*types.TagPair, error) {
	rand := fun.RandomString(RANDOM_TAG_ALPHABET, RANDOM_TAG_LENGTH)

	nonce, err := cryptag.RandomNonce()
	if err != nil {
		return nil, err
	}

	plainEnc, err := cryptag.Encrypt([]byte(plaintag), nonce, key)
	if err != nil {
		return nil, err
	}

	pair := types.NewTagPair(plainEnc, rand, nonce, plaintag)

	return pair, nil
}

func CreateTag(backend Backend, plaintag string) (*types.TagPair, error) {
	pair, err := NewTagPair(backend.Key(), plaintag)
	if err != nil {
		return nil, err
	}

	err = backend.SaveTagPair(pair)
	if err != nil {
		return nil, fmt.Errorf("Error saving tag pair: %v", err)
	}

	return pair, nil
}

func PopulateRowBeforeSave(backend Backend, row *types.Row) error {
	// Fetch all tags.  For each element of row.plainTags that doesn't
	// match an existing tag, call CreateTag().  Encrypt row.decrypted
	// and store it in row.Encrypted.  POST to server.

	// TODO: Call this in parallel with encryption below
	allTagPairs, _, err := CreateTagsFromPlain(backend, row.PlainTags())
	if err != nil {
		return fmt.Errorf("Error from CreateNewTagsFromPlain: %v", err)
	}

	// Set row.RandomTags
	for _, pair := range allTagPairs {
		if row.HasPlainTag(pair.Plain()) {
			row.RandomTags = append(row.RandomTags, pair.Random)
		}
	}

	// Set row.Encrypted

	// Could also do something like `row.Encrypt(wb.Encrypt)`
	encData, err := cryptag.Encrypt(row.Decrypted(), row.Nonce, backend.Key())
	if err != nil {
		return fmt.Errorf("Error encrypting data: %v", err)
	}
	row.Encrypted = encData

	return nil
}
