// Steve Phillips / elimisteve
// 2015.02.24

package types

import (
	"crypto/cipher"
	"time"
)

var (
	// Set by main
	SERVER_BASE_URL string
	Block           cipher.Block // Used for encryption/decryption
	Debug           bool

	// Tag-related
	RANDOM_TAG_ALPHABET = "abcdefghijklmnopqrstuvwxyz0123456789"
	RANDOM_TAG_LENGTH   = 9

	// HTTP-related
	HttpGetTimeout = 30 * time.Second
)
