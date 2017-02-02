package taber

import (
	"golang.org/x/crypto/nacl/box"
)

// Encrypt uses key material to encrypt plaintext using the NaCl SecretBox construct.
func (ks *Keys) Encrypt(plaintext, nonce []byte, to *Keys) (ciphertext []byte, err error) {
	if len(nonce) != 24 {
		return nil, ErrBadNonceLength
	}
	ciphertext = make([]byte, 0, len(plaintext)+box.Overhead)
	toArr := to.PublicArray()
	defer WipeKeyArray(toArr)
	pa := ks.PrivateArray()
	defer WipeKeyArray(pa)
	ciphertext = box.Seal(ciphertext, plaintext, nonceToArray(nonce), toArr, pa)
	return ciphertext, nil
}

// Decrypt decrypts an NaCL box to plaintext.
func (ks *Keys) Decrypt(ciphertext, nonce []byte, from *Keys) (plaintext []byte, err error) {
	var ok bool
	if len(nonce) != 24 {
		return nil, ErrBadNonceLength
	}
	plaintext = make([]byte, 0, len(ciphertext)-box.Overhead)
	fromArr := from.PublicArray()
	defer WipeKeyArray(fromArr)
	pa := ks.PrivateArray()
	defer WipeKeyArray(pa)
	plaintext, ok = box.Open(plaintext[:], ciphertext, nonceToArray(nonce), fromArr, pa)
	if !ok {
		return nil, ErrDecryptionAuthFail
	}
	return plaintext, nil
}
