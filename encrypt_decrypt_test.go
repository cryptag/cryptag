// Steve Phillips / elimisteve
// 2015.08.05

package cryptag

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	data := []byte("This is test data")
	nonce, _ := RandomNonce()
	k := []byte("012345678901234567890123456789-!")
	key, _ := ConvertKey(k)

	// Encrypt
	enc, err := Encrypt(data, nonce, key)
	if err != nil {
		t.Fatalf("Error encrypting: %v", err)
	}

	// Decrypt
	dec, err := Decrypt(enc, nonce, key)
	if err != nil {
		t.Fatalf("Error decrypting: %v", err)
	}
	if !bytes.Equal(dec, data) {
		t.Errorf("After decrypting, got\n%v\nwanted\n%v", dec, data)
	}
}
