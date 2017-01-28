package minilock

import "errors"

var (
	// ErrBadMagicBytes is returned when magic bytes didn't match expected 'miniLock'.
	ErrBadMagicBytes = errors.New("Magic bytes didn't match expected 'miniLock'")
	// ErrBadLengthPrefix is returned when header length exceeds file length.
	ErrBadLengthPrefix = errors.New("Header length exceeds file length")
	// ErrCTHashMismatch is returned when ciphertext hash did not match.
	ErrCTHashMismatch = errors.New("Ciphertext hash did not match")
	// ErrBadRecipient is returned when decryptInfo successfully decrypted but was addressed to another key.
	ErrBadRecipient = errors.New("DecryptInfo successfully decrypted but was addressed to another key")
	// ErrCannotDecrypt is returned when could not decrypt given ciphertext with given key or nonce.
	ErrCannotDecrypt = errors.New("Could not decrypt given ciphertext with given key or nonce")
	// ErrInsufficientEntropy is returned when got insufficient random bytes from RNG.
	ErrInsufficientEntropy = errors.New("Got insufficient random bytes from RNG")
	// ErrNilPlaintext is returned when got empty plaintext, can't encrypt.
	ErrNilPlaintext = errors.New("Got empty plaintext, can't encrypt")
)
