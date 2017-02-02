package taber

import (
	"bytes"
	"sync"
	"time"

	"golang.org/x/crypto/nacl/secretbox"
)

// Uses length prefixes to parse miniLock ciphertext and return a slice of
// block objects for decryption.
func walkCiphertext(ciphertext []byte) ([]block, error) {
	// Enough room for all full blocks, plus the last block, plus the name block.
	blocks := make([]block, 0, ((len(ciphertext)-256)/ConstBlockLength)+2)
	blockIndex := 0
	for loc := 0; loc < len(ciphertext); {
		//fmt.Println("walkCiphertext: loc =", loc)
		prefixLEb := ciphertext[loc : loc+4]
		prefixInt32, err := fromLittleEndian(prefixLEb)
		if err != nil {
			return nil, err
		}
		blockEnds := loc + prefixToBlockL(int(prefixInt32))
		if blockEnds > len(ciphertext) {
			return nil, ErrBadLengthPrefix
		}
		thisBlock := ciphertext[loc:blockEnds]
		blocks = append(blocks, block{Index: blockIndex, Block: thisBlock, err: nil})
		if blockEnds == len(ciphertext) {
			break
		}
		loc = blockEnds
		blockIndex = blockIndex + 1
	}
	blocks[len(blocks)-1].last = true
	return blocks, nil
}

func decryptBlock(key, baseNonce []byte, block *block) ([]byte, error) {
	var auth bool
	chunkNonce, err := makeChunkNonce(baseNonce, block.Index, block.last)
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, 0, len(block.Block)-(secretbox.Overhead+4))
	plaintext, auth = secretbox.Open(plaintext, block.Block[4:], nonceToArray(chunkNonce), keyToArray(key))
	if !auth {
		return nil, ErrBadBoxAuth
	}
	return plaintext, nil
}

// Return everything preceding the first null byte of the decrypted file-name block.
func decryptName(key, baseNonce []byte, nameBlock *block) (string, error) {
	fnBytes, err := decryptBlock(key, baseNonce, nameBlock)
	if err != nil {
		return "", err
	}
	// Trim to just the bit preceding the first null, OR the whole thing.
	fnBytes = bytes.SplitN(fnBytes, []byte{0}, 2)[0]
	return string(fnBytes), nil
}

func reassemble(plaintext []byte, chunksChan chan *enumeratedChunk, done chan bool) ([]byte, error) {
	for {
		select {
		case echunk := <-chunksChan:
			{
				if echunk.err != nil {
					return nil, echunk.err
				}
				b := echunk.beginsLocation()
				e := echunk.endsLocation()
				// End is calculated using length prefixes so must be regarded as bad
				if e > len(plaintext) {
					return nil, ErrBoxDecryptionEOP
				}
				if len(echunk.chunk) > len(plaintext[b:e]) {
					return nil, ErrBoxDecryptionEOS
				}
				copy(plaintext[b:e], echunk.chunk)
			}
		case <-done:
			{
				return plaintext, nil
			}
		default:
			{
				time.Sleep(time.Millisecond * 10)
			}
		}
	}
}

func decryptBlockAsync(key, baseNonce []byte, thisBlock *block, chunksChan chan *enumeratedChunk, wg *sync.WaitGroup) {
	// Insert decryption code here
	var echunk *enumeratedChunk
	chunk, err := decryptBlock(key, baseNonce, thisBlock)
	if err != nil {
		echunk = &enumeratedChunk{err: err, index: thisBlock.Index - 1}
	} else {
		echunk = &enumeratedChunk{index: thisBlock.Index - 1, chunk: chunk}
	}
	chunksChan <- echunk
	wg.Done()
}

// Parse blocks, fan-out using decryptBlock, re-assemble to original plaintext.
func decrypt(key, baseNonce, ciphertext []byte) (filename string, plaintext []byte, err error) {
	blocks, err := walkCiphertext(ciphertext)
	if err != nil {
		return "", nil, err
	}
	filename, err = decryptName(key, baseNonce, &blocks[0])
	if err != nil {
		return "", nil, err
	}
	chunksChan := make(chan *enumeratedChunk)
	expectedLength := 0
	wg := new(sync.WaitGroup)
	for _, thisBlock := range blocks[1:] {
		thisBlock := thisBlock
		expectedLength = expectedLength + thisBlock.ChunkLength()
		wg.Add(1)
		go decryptBlockAsync(key, baseNonce, &thisBlock, chunksChan, wg)
	}
	// If chunks are larger than the space allotted in plaintext,
	// function will throw an error.
	plaintext = make([]byte, expectedLength)
	// Translates the blocking WaitGroup into non-blocking chan bool "done".
	done := make(chan bool)
	go func(done chan bool, wg *sync.WaitGroup) {
		wg.Wait()
		done <- true
	}(done, wg)
	// Awaits chunks on chunksChan until sent on done.
	plaintext, err = reassemble(plaintext, chunksChan, done)
	if err != nil {
		return "", nil, err
	}
	return filename, plaintext, nil
}

// DecryptInfo is a structured object returned by Encrypt to go with ciphertexts, which
// provides a method for Decrypting ciphertexts. Can easily be constructed
// from raw data, passed around, serialised, etcetera.
type DecryptInfo struct {
	// Decryption key (32 bytes) and Nonce (24 bytes) required to decrypt.
	Key, BaseNonce []byte
}

// NewDecryptInfo returns a prepared DecryptInfo with a new Symmetric Key and BaseNonce.
func NewDecryptInfo() (*DecryptInfo, error) {
	key, err := makeSymmetricKey()
	if err != nil {
		return nil, err
	}
	nonce, err := makeBaseNonce()
	if err != nil {
		return nil, err
	}
	return &DecryptInfo{Key: key, BaseNonce: nonce}, nil
}

// Validate returns simply that the Key and BaseNonce look OK. It's not very clever,
// but it helps prevent accidental attempts to encrypt with a blank DecryptInfo
// created accidentally with new(DecryptInfo) or DecryptInfo{}.
func (di *DecryptInfo) Validate() bool {
	return len(di.Key) == 32 && len(di.BaseNonce) == 16
}

// Decrypt uses enclosed encryption vars to decrypt and authenticate a file.
func (di *DecryptInfo) Decrypt(ciphertext []byte) (filename string, plaintext []byte, err error) {
	if !di.Validate() {
		return "", nil, ErrBadBoxDecryptVars
	}
	return decrypt(di.Key, di.BaseNonce, ciphertext)
}
