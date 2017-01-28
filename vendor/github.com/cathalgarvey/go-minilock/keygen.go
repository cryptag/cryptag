package minilock

import (
	"github.com/cathalgarvey/go-minilock/taber"
)

// GenerateKey makes a key from an email address and passphrase, consistent
// with the miniLock algorithm. Passphrase is *not* currently checked
// for strength so it is, at present, the caller's responsibility to
// provide passphrases that don't suck!
func GenerateKey(email string, passphrase string) (*taber.Keys, error) {
	return taber.FromEmailAndPassphrase(email, passphrase)
}

// EphemeralKey generates a fully random key, usually for ephemeral uses.
func EphemeralKey() (*taber.Keys, error) {
	return taber.RandomKey()
}

// ImportID imports a miniLock ID as a public key.
func ImportID(id string) (*taber.Keys, error) {
	return taber.FromID(id)
}

// LoadKey manually loads a key from public and private binary strings.
func LoadKey(private, public []byte) *taber.Keys {
	return &taber.Keys{Private: private, Public: public}
}
