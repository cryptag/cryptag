package taber

const (
	// ConstChunkSize is the maximum length of chunks.
	ConstChunkSize = 1048576
	// ConstFilenameBlockLength is the length of the block that contains the filename,
	// also the first block in the raw ciphertext.
	ConstFilenameBlockLength = (256 + 16 + 4)
	// ConstBlockLength is the length of the encrypted chunks.
	ConstBlockLength = ConstChunkSize + 16 + 4
)
