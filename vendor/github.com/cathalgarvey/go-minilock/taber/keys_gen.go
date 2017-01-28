package taber

import (
	"bytes"
	"crypto/rand"

	"github.com/cathalgarvey/base58"
	"github.com/dchest/blake2s"
	"golang.org/x/crypto/nacl/box"
)

// Keys in Taber NaCl are stored simply as two byte slices with handy methods attached.
// They can be generated determnistically from passphrase and email, according
// to the method used by miniLock.io, or randomly generated. They can emit
// an "ID" as a string, which is the base58 representation of the public Key
// with a one-byte checksum, and can be imported from same.
// They have a "Wipe" method which should always be used when finished with the
// keys to protect keys from compromise.
type Keys struct {
	// May be empty for pubkey-only keypairs.
	Private []byte

	// Should always be full.
	Public []byte
}

// Merely verifies whether a byte slice is 32 bytes long.
func (ks *Keys) mayBeAKey(part []byte) bool {
	if part != nil && len(part) == 32 {
		return true
	}
	return false
}

// HasPublic returns whether the public component of the key appears to be a valid
// taber/nacl pubkey.
func (ks *Keys) HasPublic() bool {
	return ks.mayBeAKey(ks.Public)
}

// HasPrivate returns whether the private component of the key appears to be a valid
// taber/nacl key.
func (ks *Keys) HasPrivate() bool {
	return ks.mayBeAKey(ks.Private)
}

// PrivateArray returns a copy of the private key slice in an array. This is useful
// when some operations call for an array and not a slice. Please be aware that the
// returned array is *not* the underlying array of the private slice, it is a copy!
// This means that the Keys.Wipe() operation will not clean up this copy, so it
// should be sanitised separately if security of that grade is called for.
func (ks *Keys) PrivateArray() *[32]byte {
	if !ks.HasPrivate() {
		return nil
	}
	arr := new([32]byte)
	copy(arr[:], ks.Private)
	return arr
}

// PublicArray returns a copy of the public key in an array rather than a slice.
// Some nacl functions require arrays, not slices.
func (ks *Keys) PublicArray() *[32]byte {
	arr := new([32]byte)
	copy(arr[:], ks.Public)
	return arr
}

// PublicOnly returns a Keys object containing only the public key of this object.
func (ks *Keys) PublicOnly() *Keys {
	PK := new(Keys)
	PK.Public = make([]byte, len(ks.Public))
	copy(PK.Public, ks.Public)
	return PK
}

// RandomKey generates a fully random Keys struct from a secure random source.
func RandomKey() (*Keys, error) {
	randBytes := make([]byte, 32)
	read, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}
	if read != 32 {
		return nil, ErrInsufficientEntropy
	}
	randReader := bytes.NewReader(randBytes)
	public, private, err := box.GenerateKey(randReader)
	if err != nil {
		return nil, err
	}
	// Always be explicit about public/private material in case struct is
	// rearranged (/accidentally) later on
	return &Keys{Private: private[:], Public: public[:]}, nil
}

// FromEmailAndPassphrase generates keys using a passphrase with a GUID as salt value.
// The GUID is usually expected to be an email, but it is not normalised here, so
// be aware of the various ways one email may be represented; capitalisation,
// "+" extensions, etcetera will all result in different outcomes here.
// The passphrase is first hashed using 32-byte blake2s and is then
// passed through scrypt using the email as salt. 32 bytes of scrypt
// output are used to create a private nacl.box key and a keys object.
func FromEmailAndPassphrase(guid, passphrase string) (*Keys, error) {
	ppScrypt, err := Harden(guid, passphrase)
	if err != nil {
		return nil, err
	}
	scryptReader := bytes.NewReader(ppScrypt)
	public, private, err := box.GenerateKey(scryptReader)
	if err != nil {
		return nil, err
	}
	return &Keys{Private: private[:], Public: public[:]}, nil
}

// FromID creates a Keys struct from a checksummed ID string.
// The last byte is expected to be the blake2s checksum.
func FromID(ID string) (*Keys, error) {
	keyCSbuf, err := base58.StdEncoding.Decode([]byte(ID))
	if err != nil {
		return nil, err
	}
	if len(keyCSbuf) != 33 {
		return nil, ErrInvalidIDLength
	}
	kp := Keys{Public: keyCSbuf[:len(keyCSbuf)-1]}
	cs := keyCSbuf[len(keyCSbuf)-1:]
	// TODO: Is constant time important here at all?
	// cs2 is guaranteed length 1 here or err will be a BadProgrammingError.
	cs2, err := kp.checksum()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(cs, cs2) {
		if err != nil {
			return nil, err
		}
		return nil, ErrInvalidIDChecksum
	}
	return &kp, nil
}

// Generate 1-byte checksum using blake2s.
func (ks *Keys) checksum() ([]byte, error) {
	var written int
	blakeHasher, err := blake2s.New(&blake2s.Config{Size: 1})
	if err != nil {
		return nil, err
	}
	written, err = blakeHasher.Write(ks.Public)
	if err != nil {
		return nil, err
	}
	if written != 32 {
		return nil, ErrChecksumFail
	}
	checksum := blakeHasher.Sum(nil)
	if len(checksum) != 1 {
		return nil, ErrChecksumFail
	}
	return checksum, nil
}

// EncodeID generate base58-encoded pubkey + 1-byte blake2s checksum as a string.
func (ks *Keys) EncodeID() (string, error) {
	plen := len(ks.Public)
	idbuf := make([]byte, plen, plen+1)
	copy(idbuf, ks.Public)
	cs, err := ks.checksum()
	if err != nil {
		return "", err
	}
	idbuf = append(idbuf, cs[0])
	id := base58.StdEncoding.Encode(idbuf)
	return string(id), nil
}

// Wipe overwrites memory containing key material; calling this method when
// finished with a key is strongly advised to prevent compromise.
func (ks *Keys) Wipe() (err error) {
	if ks.HasPrivate() {
		err = wipeByteSlice(ks.Private)
		if err != nil {
			return err
		}
	}
	return wipeByteSlice(ks.Public)
}
