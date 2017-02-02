package taber

// Whereas block represents a section of ciphertext, enumeratedChunk represents
// a section of plaintext.
type enumeratedChunk struct {
	index int
	chunk []byte
	err   error
}

func (enChk *enumeratedChunk) beginsLocation() int {
	return enChk.index * ConstChunkSize
}

func (enChk *enumeratedChunk) endsLocation() int {
	return enChk.beginsLocation() + len(enChk.chunk)
}
