// Steve Phillips / elimisteve
// 2015.02.24

package types

import (
	"fmt"
	"log"
	"strings"

	"github.com/thecloakproject/utils/crypt"
)

type TagPair struct {
	PlainEncrypted []byte `json:"plain_encrypted"`
	Random         string `json:"random"`

	plain string
}

func (pair *TagPair) setUnexported() error {
	log.Printf("\nPlainEncrypted: %v (%s)\n", pair.PlainEncrypted, pair.PlainEncrypted)
	plain, err := crypt.AESDecryptBytes(Block, pair.PlainEncrypted)
	if err != nil {
		return fmt.Errorf("Error decrypting plain tag: %v", err)
	}

	log.Printf("plain == %v\n", plain)
	pair.plain = strings.TrimSpace(string(plain))
	log.Printf("string(plain) == %s\n\n", plain)

	return nil
}

// TagPairs

type TagPairs []*TagPair

func (pairs TagPairs) String() string {
	var s string
	for _, pair := range pairs {
		s += fmt.Sprintf("%#v\n", pair)
	}
	return s
}

func (pairs TagPairs) AllRandom() []string {
	random := make([]string, 0, len(pairs))
	for _, p := range pairs {
		random = append(random, p.Random)
	}
	return random
}

func (pairs TagPairs) AllPlain() []string {
	plain := make([]string, 0, len(pairs))
	for _, p := range pairs {
		plain = append(plain, p.plain)
	}
	return plain
}

func (pairs TagPairs) FilterByRandomTags(random []string) (filtered TagPairs) {
	filtered = pairs[:] // TODO: Ensure this is good enough
	for _, randtag := range random {
		filtered = filtered.FilterByRandomTag(randtag)
	}
	return filtered
}

func (pairs TagPairs) FilterByRandomTag(randtag string) (filtered TagPairs) {
	filtered = pairs[:] // TODO: Ensure this is good enough
	for _, pair := range pairs {
		log.Printf("Does `%s` == `%s`?\n", pair.Random, randtag)
		if pair.Random == randtag {
			log.Printf("Yes!\n")
			filtered = append(filtered, pair)
		} else {
			log.Printf("No!\n")
		}
	}
	return filtered
}

func (pairs TagPairs) FilterByPlainTags(plaintags []string) (filtered TagPairs) {
	filtered = pairs[:] // TODO: Ensure this is good enough
	for _, plain := range plaintags {
		filtered = filtered.FilterByPlainTag(plain)
	}
	return filtered
}

func (pairs TagPairs) FilterByPlainTag(plain string) (filtered TagPairs) {
	filtered = pairs[:] // TODO: Ensure this is good enough
	for _, pair := range pairs {
		if pair.plain == plain {
			filtered = append(filtered, pair)
		}
	}
	return filtered
}

func (pairs TagPairs) setUnexported() error {
	var err error
	for _, pair := range pairs {
		if err = pair.setUnexported(); err != nil {
			return err
		}
	}
	return nil
}
