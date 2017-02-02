package taber

import "errors"

var (
	// ErrBadKeyLength is returned when encryption key must be 32 bytes long.
	ErrBadKeyLength = errors.New("Encryption key must be 32 bytes long")
	// ErrBadBaseNonceLength is returned when length of base_nonce must be 16.
	ErrBadBaseNonceLength = errors.New("Length of base_nonce must be 16")
	// ErrBadLengthPrefix is returned when block length prefixes indicate a length longer than the remaining ciphertext.
	ErrBadLengthPrefix = errors.New("Block length prefixes indicate a length longer than the remaining ciphertext")
	// ErrBadPrefix is returned when chunk length prefix is longer than 4 bytes, would clobber ciphertext.
	ErrBadPrefix = errors.New("Chunk length prefix is longer than 4 bytes, would clobber ciphertext")
	// ErrBadBoxAuth is returned when authentication of box failed on opening.
	ErrBadBoxAuth = errors.New("Authentication of box failed on opening")
	// ErrBadBoxDecryptVars is returned when key or Nonce is not correct length to attempt decryption.
	ErrBadBoxDecryptVars = errors.New("Key or Nonce is not correct length to attempt decryption")
	// ErrBoxDecryptionEOP is returned when declared length of chunk would write past end of plaintext slice!.
	ErrBoxDecryptionEOP = errors.New("Declared length of chunk would write past end of plaintext slice!")
	// ErrBoxDecryptionEOS is returned when chunk length is longer than expected slot in plaintext slice.
	ErrBoxDecryptionEOS = errors.New("Chunk length is longer than expected slot in plaintext slice")
	// ErrFilenameTooLong is returned when filename cannot be longer than 256 bytes.
	ErrFilenameTooLong = errors.New("Filename cannot be longer than 256 bytes")
	// ErrNilPlaintext is returned when asked to encrypt empty plaintext.
	ErrNilPlaintext = errors.New("Asked to encrypt empty plaintext")
)
