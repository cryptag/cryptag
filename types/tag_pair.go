// Steve Phillips / elimisteve
// 2015.03.01

package types

import (
	"errors"
	"fmt"
)

var (
	ErrTagPairNotFound = errors.New("TagPair(s) not found")
)

type TagPair struct {
	PlainEncrypted []byte    `json:"plain_encrypted"`
	Random         string    `json:"random"`
	Nonce          *[24]byte `json:"nonce"`

	plain string
}

func NewTagPair(plainEnc []byte, random string, nonce *[24]byte, plaintag string) *TagPair {
	return &TagPair{
		PlainEncrypted: plainEnc,
		Random:         random,
		Nonce:          nonce,
		plain:          plaintag,
	}
}

func (pair *TagPair) Plain() string {
	return pair.plain
}

// Decrypt sets pair.plain based off of pair.PlainEncrypted
func (pair *TagPair) Decrypt(decrypt func(enc []byte, nonce *[24]byte) ([]byte, error)) error {
	plain, err := decrypt(pair.PlainEncrypted, pair.Nonce)
	if err != nil {
		return fmt.Errorf("Error decrypting plain tag `%s` (%v): %v",
			pair.PlainEncrypted, pair.PlainEncrypted, err)
	}

	pair.plain = string(plain)

	return nil
}
