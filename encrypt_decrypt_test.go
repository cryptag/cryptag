// Steve Phillips / elimisteve
// 2015.08.05

package cryptag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	plain := []byte("Original plaintext to be encrypted then decrypted")
	nonce, _ := RandomNonce()
	key, _ := ConvertKey([]byte("012345678901234567890123456789-!"))

	// Encrypt
	enc, err := Encrypt(plain, nonce, key)
	if err != nil {
		t.Fatalf("Error encrypting: %v", err)
	}

	// Decrypt
	dec, err := Decrypt(enc, nonce, key)
	if err != nil {
		t.Fatalf("Error decrypting: %v", err)
	}

	assert.Equal(t, dec, plain, "Decrypted data doesn't match original plaintext")
}
