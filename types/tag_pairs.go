// Steve Phillips / elimisteve
// 2015.02.24

package types

import (
	"bytes"
	"fmt"

	"github.com/thecloakproject/utils/crypt"
)

type TagPair struct {
	PlainEncrypted []byte `json:"plain_encrypted"`
	Random         string `json:"random"`

	plain string
}

// setUnexported sets pair.plain based off of pair.PlainEncrypted
func (pair *TagPair) setUnexported() error {
	plain, err := crypt.AESDecryptBytes(Block, pair.PlainEncrypted)
	if err != nil {
		return fmt.Errorf("Error decrypting plain tag: %v", err)
	}
	// TODO: crypt.AESDecryptBytes should do this for me unless some
	// trailing '\x00's are valid
	plain = bytes.TrimRight(plain, "\x00")

	pair.plain = string(plain)

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

func (pairs TagPairs) AllPlain() []string {
	plain := make([]string, 0, len(pairs))
	for _, p := range pairs {
		plain = append(plain, p.plain)
	}
	return plain
}

func (pairs TagPairs) HaveAllPlainTags(plaintags []string) TagPairs {
	var matches TagPairs
	for _, pair := range pairs {
		for _, plain := range plaintags {
			if pair.plain == plain {
				matches = append(matches, pair)
				break
			}
		}
	}
	return matches
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
