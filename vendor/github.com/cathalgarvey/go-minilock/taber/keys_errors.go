package taber

import "errors"

var (
	// ErrChecksumFail is returned when generating checksum value failed.
	ErrChecksumFail = errors.New("Generating checksum value failed")
	// ErrInsufficientEntropy is returned when got insufficient random bytes from RNG.
	ErrInsufficientEntropy = errors.New("Got insufficient random bytes from RNG")
	// ErrInvalidIDLength is returned when provided public ID was not expected length (33 bytes when decoded).
	ErrInvalidIDLength = errors.New("Provided public ID was not expected length (33 bytes when decoded)")
	// ErrInvalidIDChecksum is returned when provided public ID had an invalid checksum.
	ErrInvalidIDChecksum = errors.New("Provided public ID had an invalid checksum")
	// ErrPrivateKeyOpOnly is returned when cannot conduct specified operation using a public-only keypair.
	ErrPrivateKeyOpOnly = errors.New("Cannot conduct specified operation using a public-only keypair")
	// ErrBadNonceLength is returned when nonce length must be 24 length.
	ErrBadNonceLength = errors.New("Nonce length must be 24 length")
	// ErrDecryptionAuthFail is returned when authentication of decryption using keys failed.
	ErrDecryptionAuthFail = errors.New("Authentication of decryption using keys failed")
)
