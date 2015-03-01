// Steve Phillips / elimisteve
// 2015.03.01

package types

import (
	"bytes"
	"crypto/cipher"
	"fmt"

	"github.com/thecloakproject/utils/crypt"
)

type TagPair struct {
	PlainEncrypted []byte `json:"plain_encrypted"`
	Random         string `json:"random"`

	plain string
}

func NewTagPair(plainEnc []byte, random, plaintag string) *TagPair {
	return &TagPair{
		PlainEncrypted: plainEnc,
		Random:         random,
		plain:          plaintag,
	}
}

func (pair *TagPair) Plain() string {
	return pair.plain
}

// Decrypt sets pair.plain based off of pair.PlainEncrypted
func (pair *TagPair) Decrypt(block cipher.Block) error {
	plain, err := crypt.AESDecryptBytes(block, pair.PlainEncrypted)
	if err != nil {
		return fmt.Errorf("Error decrypting plain tag: %v", err)
	}
	// TODO: crypt.AESDecryptBytes should do this for me unless some
	// trailing '\x00's are valid
	plain = bytes.TrimRight(plain, "\x00")

	pair.plain = string(plain)

	return nil
}
