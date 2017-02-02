package taber

import (
	"sync"
	"time"

	"golang.org/x/crypto/nacl/secretbox"
)

func makeChunkNonce(base_nonce []byte, chunk_number int, last bool) ([]byte, error) {
	if len(base_nonce) != 16 {
		return nil, ErrBadBaseNonceLength
	}
	n := make([]byte, len(base_nonce)+8)
	chunk_num_b, err := toLittleEndian(int32(chunk_number))
	if err != nil {
		return nil, err
	}
	copy(n, base_nonce)
	copy(n[len(base_nonce):], chunk_num_b)
	if last {
		n[len(n)-1] = n[len(n)-1] | 128
	}
	return n, nil
}

func encryptChunk(key, base_nonce, chunk []byte, index int, last bool) (*block, error) {
	// Handling of last-chunk is done using the chunk nonce,
	chunk_nonce, err := makeChunkNonce(base_nonce, index, last)
	if err != nil {
		return nil, err
	}
	// Make room for the 4-byte length prefix, the ciphertext, plus the ciphertext overhead.
	ciphertext := make([]byte, 4, len(chunk)+secretbox.Overhead+4)
	// Prepend length.
	bl_len, err := toLittleEndian(int32(len(chunk)))
	if err != nil {
		return nil, err
	}
	if len(bl_len) > 4 {
		return nil, ErrBadPrefix
	}
	copy(ciphertext, bl_len)
	// Put the ciphertext in the space after the first four bytes.
	ciphertext = secretbox.Seal(ciphertext, chunk, nonceToArray(chunk_nonce), keyToArray(key))
	// Get 4-byte length prefix, verify it's the right length JIC.
	return &block{Block: ciphertext, Index: index}, nil
}

// Convenience for the special case.
func prepareNameChunk(filename string) ([]byte, error) {
	// Prepare name
	fn_bytes := []byte(filename)
	if len(fn_bytes) > 256 {
		return nil, ErrFilenameTooLong
	}
	padded_name := make([]byte, 256, 256)
	copy(padded_name, fn_bytes)
	return padded_name, nil
}

// Chunk up a file and encrypt each chunk separately, returning each chunk through
// block_chan for reassembly. Blocks can and will arrive out of order through block_chan.
func encryptToChan(filename string, key, base_nonce, file_data []byte, block_chan chan *block, done chan bool) (err error) {
	if base_nonce == nil {
		base_nonce, err = makeBaseNonce()
		if err != nil {
			return err
		}
	}
	filename_chunk, err := prepareNameChunk(filename)
	if err != nil {
		return err
	}
	fn_block, err := encryptChunk(key, base_nonce, filename_chunk, 0, false)
	if err != nil {
		return err
	}
	block_chan <- fn_block
	// Get expected chunk number so special treatment of last chunk can be done
	// correctly.
	num_chunks := numChunks(len(file_data), ConstChunkSize)
	wg := new(sync.WaitGroup)
	for i, chunk := range chunkify(file_data, ConstChunkSize) {
		block_number := i + 1
		wg.Add(1)
		// Fan out the job of encrypting each chunk. Each ciphertext block gets passed
		// back through block_chan. WaitGroup wg makes sure all goroutines are finished
		// prior to passing back "done".
		go func(key, base_nonce, chunk []byte, block_number int, block_chan chan *block, wg *sync.WaitGroup) {
			var (
				ciphertext *block
				err        error
			)
			if block_number == num_chunks {
				ciphertext, err = encryptChunk(key, base_nonce, chunk, block_number, true)
			} else {
				ciphertext, err = encryptChunk(key, base_nonce, chunk, block_number, false)
			}
			if err != nil {
				ciphertext.err = err
			}
			block_chan <- ciphertext
			wg.Done()
		}(key, base_nonce, chunk, block_number, block_chan, wg)

	}
	wg.Wait()
	done <- true
	return nil
}

// Adds "base_nonce" to public facing version for testing purposes.
func encrypt(filename string, key, base_nonce, file_data []byte) (ciphertext []byte, err error) {
	if len(key) != 32 {
		return nil, ErrBadKeyLength
	}
	// Pre-allocate space to help assemble the ciphertext afterwards..
	num_chunks := numChunks(len(file_data), ConstChunkSize)
	// Now allocate all but the last block. The last block is *appended* to the
	// output, the others are *copied*.
	// Each block requires 4 for the LE int length prefix, CHUNK_SIZE for the block,
	// and 16 for the encryption overhead.
	max_length := ConstFilenameBlockLength + (num_chunks * ConstBlockLength)
	ciphertext = make([]byte, max_length-ConstBlockLength, max_length)
	// Now fan-out the job of encrypting the file...
	block_chan := make(chan *block)
	done_chan := make(chan bool)
	go encryptToChan(filename, key, base_nonce, file_data, block_chan, done_chan)
	// And then fan-in re-assembly.
	for {
		select {
		case this_block := <-block_chan:
			{
				if this_block.err != nil {
					// Blocks may return errors, cancel on any error.
					// This leaves the goroutines running but they ought to run out on their own?
					return nil, this_block.err
				}
				if this_block.Index > 0 && this_block.Index < num_chunks {
					// Find correct location for each chunk and copy in.
					begins := this_block.BeginsLocation()
					ends := begins + len(this_block.Block)
					copy(ciphertext[begins:ends], this_block.Block)
				} else if this_block.Index == num_chunks {
					ciphertext = append(ciphertext, this_block.Block...)
				} else {
					// Filename chunk, special case.
					copy(ciphertext, this_block.Block)
				}
			}
		case <-done_chan:
			{
				goto encrypt_finished
			}
		default:
			{
				time.Sleep(time.Millisecond * 10)
			}
		}
	}
encrypt_finished:
	return ciphertext, nil
}

// Encrypt symmetrically using this DecryptInfo object.
func (self *DecryptInfo) Encrypt(filename string, file_data []byte) (ciphertext []byte, err error) {
	if file_data == nil || len(file_data) == 0 {
		return nil, ErrNilPlaintext
	}
	ciphertext, err = encrypt(filename, self.Key, self.BaseNonce, file_data)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// Generates random key/nonce, encrypts the data with it, and returns a DecryptInfo
// object for storage, serialisation or deconstruction, along with the ciphertext.
// According to the miniLock encryption protocol, the filename is encrypted in the
// first block.
func Encrypt(filename string, file_data []byte) (DI *DecryptInfo, ciphertext []byte, err error) {
	DI = new(DecryptInfo)
	DI.Key, err = makeSymmetricKey()
	if err != nil {
		return nil, nil, err
	}
	DI.BaseNonce, err = makeBaseNonce()
	if err != nil {
		return nil, nil, err
	}
	ciphertext, err = DI.Encrypt(filename, file_data)
	if err != nil {
		return nil, nil, err
	}
	return DI, ciphertext, nil
}
