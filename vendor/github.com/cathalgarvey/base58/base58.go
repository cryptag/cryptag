// Package base58 implements a human-friendly base58 encoding.
//
// As opposed to base64 and friends, base58 is typically used to
// convert integers. You can use big.Int.SetBytes to convert arbitrary
// bytes to an integer first, and big.Int.Bytes the other way around.
package base58

import (
	"math/big"
	"strconv"
)

const caps_last_alphabet = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
const caps_first_alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Uses alphabet "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
// Default for compatibility reasons. The "EncodeBig" and "DecodeToBig"
// top-level functions are just shortcuts for methods on this encoding.
var DefaultEncoding = NewEncoder(caps_last_alphabet)

// Uses alphabet "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
// Standard because it appears to be in wider use, v/v miniLock, Bitcoin,
// Python's base58, etc.
var StdEncoding = NewEncoder(caps_first_alphabet)

// Encode encodes src, appending to dst. Be sure to use the returned
// new value of dst.
// This is a shortcut to DefaultEncoding.EncodeBig. It's suggested for
// maintainability reasons to use the full reference in new code.
var EncodeBig func([]byte, *big.Int)[]byte = DefaultEncoding.EncodeBig

// Decode a big integer from the bytes. Returns an error on corrupt
// input.
// This is a shortcut to DefaultEncoding.DecodeToBig. It's suggested for
// maintainability reasons to use the full reference in new code.
var DecodeToBig func([]byte)(*big.Int,error) = DefaultEncoding.DecodeToBig

// Functions normally global re-implemented as shortcuts to library-level
// encodingAlphabets, mimicing the behaviour of base64/base32 and providing
// more flexibility while retaining backwards compatibility.
// Users can define their own alphabets, two are provided.
type encodingAlphabet struct{
	alphabet string
	decodeMap [256]byte
}

// If a custom alphabet is desired it can be defined here by passing a length 62
// string of unique characters. Two encodings are provided in this library already,
// representing "caps first" or "caps last" base58.
func NewEncoder(alphabet string) *encodingAlphabet {
  enc := new(encodingAlphabet)
 	enc.alphabet = alphabet
  // Is this necessary after `new`?
	for i := 0; i < len(enc.decodeMap); i++ {
		enc.decodeMap[i] = 0xFF
	}
	for i := 0; i < len(enc.alphabet); i++ {
		enc.decodeMap[enc.alphabet[i]] = byte(i)
	}
  return enc
}

// Decode a big integer from the bytes. Returns an error on corrupt
// input.
func (self *encodingAlphabet) DecodeToBig(src []byte) (*big.Int, error) {
	n := new(big.Int)
	radix := big.NewInt(58)
	for i := 0; i < len(src); i++ {
		b := self.decodeMap[src[i]]
		if b == 0xFF {
			return nil, CorruptInputError(i)
		}
		n.Mul(n, radix)
		n.Add(n, big.NewInt(int64(b)))
	}
	return n, nil
}

// Encode encodes src, appending to dst. Be sure to use the returned
// new value of dst.
func (self *encodingAlphabet) EncodeBig(dst []byte, src *big.Int) []byte {
	start := len(dst)
	n := new(big.Int)
	n.Set(src)
	radix := big.NewInt(58)
	zero := big.NewInt(0)

	for n.Cmp(zero) > 0 {
		mod := new(big.Int)
		n.DivMod(n, radix, mod)
		dst = append(dst, self.alphabet[mod.Int64()])
	}

	for i, j := start, len(dst)-1; i < j; i, j = i+1, j-1 {
		dst[i], dst[j] = dst[j], dst[i]
	}
	return dst
}

// Encode a byte string to binary base58 using this encoding.
func (self *encodingAlphabet) Encode(data []byte) []byte {
  bigInt := big.NewInt(0).SetBytes(data)
  return self.EncodeBig(nil, bigInt)
}

// Decode binary base58 to a byte string using this encoding.
func (self *encodingAlphabet) Decode(enc []byte) ([]byte, error) {
  bigInt, err := self.DecodeToBig(enc)
  if err != nil {
    return nil, err
  }
  return bigInt.Bytes(), nil
}

type CorruptInputError int64

func (e CorruptInputError) Error() string {
	return "illegal base58 data at input byte " + strconv.FormatInt(int64(e), 10)
}
