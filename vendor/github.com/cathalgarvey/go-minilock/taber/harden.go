package taber

import (
	"github.com/dchest/blake2s"
	"golang.org/x/crypto/scrypt"
)

func Harden(salt, passphrase string) ([]byte, error) {
	pp_blake := blake2s.Sum256([]byte(passphrase))
	return scrypt.Key(pp_blake[:], []byte(salt), 131072, 8, 1, 32)
}
