// Steve Phillips / elimisteve
// 2015.03.01

package backend

import (
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
	Encrypt(plain []byte, nonce *[24]byte) ([]byte, error)
	Decrypt(cipher []byte, nonce *[24]byte) ([]byte, error)

	AllTagPairs() (types.TagPairs, error)
	TagPairsFromRandomTags(randtags []string) (types.TagPairs, error)
	SaveTagPair(*types.TagPair) (*types.TagPair, error)

	RowsFromPlainTags(plaintags []string) (types.Rows, error)
	SaveRow(*types.Row) (*types.Row, error)
	DeleteRows(randTags []string) error
}

func randomFromPlain(backend Backend, plaintags []string) ([]string, error) {
	// Get encrypted JSON containing the tag pairs, decrypt JSON,
	// unmarshal.  for p in plain, find *TagPair with pair.Plain() ==
	// plain, append pair.Random to results

	pairs, err := backend.AllTagPairs()
	if err != nil {
		return nil, fmt.Errorf("Error from AllTagPairs: %v", err)
	}

	if types.Debug {
		log.Printf("%d pairs fetched from AllTagPairs: ", len(pairs))
		for _, pair := range pairs {
			log.Printf("  * %#v\n", pair)
		}
	}

	var randoms []string
	for _, plain := range plaintags {
		for _, pair := range pairs {
			if plain == pair.Plain() {
				randoms = append(randoms, pair.Random)
				break
			}
		}
	}

	if types.Debug && len(plaintags) != len(randoms) {
		log.Printf("Mapped plain `%#v` to random `%#v`\n", plaintags, randoms)
	}

	if len(randoms) == 0 {
		return nil, types.ErrTagPairNotFound
	}

	return randoms, nil
}

func CreateTagsFromPlain(backend Backend, plaintags []string) (allPairs types.TagPairs, newPairs types.TagPairs, err error) {
	// Fetch all tags
	allPairs, err = backend.AllTagPairs()
	if err != nil {
		return nil, nil, fmt.Errorf("Error from AllTagPairs: %v", err)
	}

	log.Printf("Fetched all %d TagPairs from server\n", len(allPairs))

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

func CreateTag(backend Backend, plaintag string) (*types.TagPair, error) {
	rand := fun.RandomString(RANDOM_TAG_ALPHABET, RANDOM_TAG_LENGTH)

	nonce, err := cryptag.RandomNonce()
	if err != nil {
		return nil, err
	}

	plainEnc, err := backend.Encrypt([]byte(plaintag), nonce)
	if err != nil {
		return nil, err
	}

	pair := types.NewTagPair(plainEnc, rand, nonce, plaintag)
	return backend.SaveTagPair(pair)
}

func PopulateRowBeforeSave(backend Backend, row *types.Row) (*types.Row, error) {
	// Fetch all tags.  For each element of row.plainTags that doesn't
	// match an existing tag, call CreateTag().  Encrypt row.decrypted
	// and store it in row.Encrypted.  POST to server.

	// TODO: Call this in parallel with encryption below
	allTagPairs, _, err := CreateTagsFromPlain(backend, row.PlainTags())
	if err != nil {
		return nil, fmt.Errorf("Error from CreateNewTagsFromPlain: %v", err)
	}

	// Set row.RandomTags
	for _, pair := range allTagPairs {
		if row.HasPlainTag(pair.Plain()) {
			row.RandomTags = append(row.RandomTags, pair.Random)
		}
	}

	// Set row.Encrypted

	// Could also do something like `row.Encrypt(wb.Encrypt)`
	encData, err := backend.Encrypt(row.Decrypted(), row.Nonce)
	if err != nil {
		return nil, fmt.Errorf("Error encrypting data: %v", err)
	}
	row.Encrypted = encData

	return row, nil
}

func PopulateRowAfterGet(backend Backend, row *types.Row) error {
	if err := row.Decrypt(backend.Decrypt); err != nil {
		return err
	}
	if err := row.SetPlainTags(backend.TagPairsFromRandomTags); err != nil {
		return err
	}
	return nil
}
