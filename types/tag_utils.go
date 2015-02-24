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

func FetchByPlainTags(plaintags []string) (Rows, error) {
	randtags, err := RandomFromPlain(plaintags)
	if err != nil {
		return nil, fmt.Errorf("Error from RandomTagsFromPlain: %v", err)
	}
	if Debug {
		log.Printf("After RandomTagsFromPlain: randtags == `%#v`\n", randtags)
	}

	fullURL := SERVER_BASE_URL + "?tags=" + strings.Join(randtags, ",")
	if Debug {
		log.Printf("fullURL == `%s`\n", fullURL)
	}

	rows, err := GetRowsFrom(fullURL)
	if err != nil {
		return nil, fmt.Errorf("Error from GetRowsFrom: %v", err)
	}

	return rows, nil
}

func RandomFromPlain(plaintags []string) ([]string, error) {
	// Download encrypted JSON containing the tag pairs, decrypt JSON,
	// unmarshal.  for p in plain, find *TagPair with pair.plain ==
	// plain, append pair.Random to results

	pairs, err := GetTagPairs()
	if err != nil {
		return nil, fmt.Errorf("Error from GetTagPairs: %v", err)
	}

	if Debug {
		log.Printf("%d pairs fetched from GetTagPairs: ", len(pairs))
		for _, pair := range pairs {
			log.Printf("  * %#v\n", pair)
		}
	}

	var randoms []string
	for _, plain := range plaintags {
		for _, pair := range pairs {
			if pair.plain == plain {
				randoms = append(randoms, pair.Random)
				break
			}
		}
	}

	if Debug && len(plaintags) != len(randoms) {
		log.Printf("Mapped plain `%#v` to random `%#v`\n", plaintags, randoms)
	}

	return randoms, nil
}

func GetTagPairsFromRandom(randtags []string) (TagPairs, error) {
	if len(randtags) == 0 {
		return nil, fmt.Errorf("Can't get 0 tags")
	}

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
				log.Printf("Error calling CreateTag(%q): %v\n", plain, err)
				continue
			}
			if Debug {
				log.Printf("Created tag pair `%#v` (%p)\n", pair, pair)
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

	if Debug {
		log.Printf("POSTing tag pair data: `%s`\n", pairBytes)
	}

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

		err2 := rows.setUnexported()
		if err2 != nil {
			log.Printf("setUnexported err2: %v\n", err2)
		}

		return nil, fmt.Errorf("Got HTTP %d from server for data: `%s`",
			resp.StatusCode, rows)
	}

	if Debug {
		log.Printf("New *TagPair created: `%#v`\n", pair)
	}

	return pair, nil
}

func GetTagPairs() (TagPairs, error) {
	return GetTagsFrom(SERVER_BASE_URL + "/tags")
}
