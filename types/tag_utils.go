// Steve Phillips / elimisteve
// 2015.02.24

package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/222Labs/help"
	"github.com/elimisteve/fun"
	"github.com/thecloakproject/utils/crypt"
)

func FetchByPlainTags(url string, plain []string) (Rows, error) {
	randtags, err := RandomTagsFromPlain(plain)
	if err != nil {
		return nil, fmt.Errorf("Error from RandomTagsFromPlain: %v", err)
	}
	log.Printf("After RandomTagsFromPlain: randtags == `%v`\n", randtags)

	fullURL := url + "?tags=" + strings.Join(randtags, "+")
	rows, err := GetRowsFrom(fullURL)
	if err != nil {
		return nil, fmt.Errorf("Error from GetRowsFrom: %v", err)
	}

	return rows, nil
}

func RandomTagsFromPlain(plain []string) ([]string, error) {
	// Download encrypted JSON containing the tag pairs, decrypt JSON,
	// unmarshal.  for p in plain, find *TagPair with pair.plain ==
	// plain, append pair.Random to results

	pairs, err := GetTagPairs()
	if err != nil {
		return nil, fmt.Errorf("Error from GetTagPairs: %v", err)
	}
	log.Printf("pairs fetched from GetTagPairs: ")
	for _, pair := range pairs {
		log.Printf("  * %#v\n", pair)
	}

	random := pairs.FilterByPlainTags(plain).AllRandom()

	return random, nil
}

func GetTagPairsFromRandom(randtags ...string) (TagPairs, error) {
	if len(randtags) == 0 {
		return nil, fmt.Errorf("Can't get 0 tags")
	}

	// TODO: Implement this. Currently simulates a remote
	// lookup. Should GET, unmarshal to Rows, probably unmarshal each
	// row.Data to a custom data type, then... ? It's late...

	return GetTagsFrom(SERVER_BASE_URL + "/tags?tags=" + strings.Join(randtags, "+"))
}

// GetTagsFrom fetches the encrypted tag pairs at url, decrypts them,
// and unmarshals them into a TagPairs value
func GetTagsFrom(url string) (TagPairs, error) {
	var pairs TagPairs
	if err := fun.FetchInto(url, HttpGetTimeout, &pairs); err != nil {
		return nil, fmt.Errorf("Error from FetchInto: %v", err)
	}

	if err := pairs.setUnexported(); err != nil {
		return nil, fmt.Errorf("Error setting unexported TagPair fields: %v", err)
	}

	return pairs, nil
}

func GetTagPairFromRandom(randtag string) (*TagPair, error) {
	pairs, err := GetTagPairsFromRandom(randtag)
	if err != nil {
		return nil, err
	}

	if len(pairs) != 1 {
		return nil, fmt.Errorf("GetTags returned %d results, wanted 1", len(pairs))
	}

	return pairs[0], nil
}

//
// Create
//

func CreateTagsFromPlain(plaintags []string) (allPairs TagPairs, newPairs TagPairs, err error) {
	// Fetch all tags
	allPairs, err = GetTagPairs()
	if err != nil {
		return nil, nil, fmt.Errorf("Error from GetTags: %v", err)
	}

	// Find out which members of plaintags don't have an existing,
	// corresponding TagPair

	existingPlain := allPairs.AllPlain()

	for _, plain := range plaintags {
		if !fun.SliceContains(existingPlain, plain) {
			// TODO: Parallelize
			pair, err := CreateTag(plain)
			if err != nil {
				continue
			}
			newPairs = append(newPairs, pair)
		}
	}

	allPairs = append(allPairs, newPairs...)

	return allPairs, newPairs, nil
}

func CreateTag(plaintag string) (*TagPair, error) {
	rand := fun.RandomString(RANDOM_TAG_ALPHABET, RANDOM_TAG_LENGTH)

	plainEnc, err := crypt.AESEncryptBytes(Block, []byte(plaintag))
	if err != nil {
		return nil, err
	}

	pair := &TagPair{PlainEncrypted: plainEnc, plain: plaintag, Random: rand}
	pairBytes, _ := json.Marshal(pair)

	resp, err := http.Post(SERVER_BASE_URL+"/tags", "application/json",
		bytes.NewReader(pairBytes))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		var rows Rows
		if err = help.ReadInto(resp.Body, &rows); err != nil {
			return nil, fmt.Errorf("Got HTTP %d from server. Error decoding resp: %v",
				resp.StatusCode, err)
		}
		_ = rows.setUnexported()

		return nil, fmt.Errorf("Got HTTP %d from server for data: `%s`",
			resp.StatusCode, rows)
	}

	log.Printf("New *TagPair created: `%#v`\n", pair)

	return pair, nil
}

func GetTagPairs() (TagPairs, error) {
	return GetTagsFrom(SERVER_BASE_URL + "/tags")
}
