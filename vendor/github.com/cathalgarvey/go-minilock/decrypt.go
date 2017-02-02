package minilock

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"

	"github.com/cathalgarvey/go-minilock/taber"
	"github.com/dchest/blake2s"
)

// ParseFile opens a file and passes to ParseFileContents
func ParseFile(filepath string) (header *miniLockv1Header, ciphertext []byte, err error) {
	fc, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, nil, err
	}
	return ParseFileContents(fc)
}

// ParseFileContents parses a miniLock file and returns header and ciphertext.
func ParseFileContents(contents []byte) (header *miniLockv1Header, ciphertext []byte, err error) {
	var (
		headerLengthi32 int32
		headerLength    int
		headerBytes     []byte
	)
	if string(contents[:8]) != magicBytes {
		return nil, nil, ErrBadMagicBytes
	}
	headerLengthi32, err = fromLittleEndian(contents[8:12])
	if err != nil {
		return nil, nil, err
	}
	headerLength = int(headerLengthi32)
	if 12+headerLength > len(contents) {
		return nil, nil, ErrBadLengthPrefix
	}
	headerBytes = contents[12 : 12+headerLength]
	ciphertext = contents[12+headerLength:]
	header = new(miniLockv1Header)
	err = json.Unmarshal(headerBytes, header)
	if err != nil {
		return nil, nil, err
	}
	return header, ciphertext, nil
}

// DecryptFileContents parses header and ciphertext from a file, decrypts the
// header with recipientKey, and uses details therein to decrypt the enclosed file.
// Returns sender, filename, file contents if successful, or an error if not;
// Check the error to see if it's benign (cannot decrypt with given key) or bad.
func DecryptFileContents(fileContents []byte, recipientKey *taber.Keys) (senderID, filename string, contents []byte, err error) {
	var (
		header     *miniLockv1Header
		ciphertext []byte
	)
	header, ciphertext, err = ParseFileContents(fileContents)
	if err != nil {
		return "", "", nil, err
	}
	return header.DecryptContents(ciphertext, recipientKey)
}

// DecryptFileContentsWithStrings is the highest-level API for decryption.
// It uses the recipient's email and passphrase to generate their key, attempts
// decryption, and wipes keys when finished.
func DecryptFileContentsWithStrings(fileContents []byte, recipientEmail, recipientPassphrase string) (senderID, filename string, contents []byte, err error) {
	var recipientKey *taber.Keys
	recipientKey, err = taber.FromEmailAndPassphrase(recipientEmail, recipientPassphrase)
	if err != nil {
		return
	}
	defer recipientKey.Wipe()
	return DecryptFileContents(fileContents, recipientKey)
}

// DecryptFile - Given a ciphertext, walk it into length prefixed chunks and decrypt/reassemble
// each chunk, then validate the hash of the file against the hash given in FileInfo.
// The result is a validated, decrypted filename and file contents byte-slice.
func (fi *FileInfo) DecryptFile(ciphertext []byte) (filename string, contents []byte, err error) {
	var (
		hash [32]byte
		DI   taber.DecryptInfo
	)
	hash = blake2s.Sum256(ciphertext)
	if !bytes.Equal(fi.FileHash, hash[:]) {
		return "", nil, ErrCTHashMismatch
	}
	DI = taber.DecryptInfo{Key: fi.FileKey, BaseNonce: fi.FileNonce}
	return DI.Decrypt(ciphertext)
}

// DecryptDecryptInfo is used to extract a decryptInfo object by attempting decryption
// with a given recipientKey. This must be attempted for each decryptInfo in the header
// until one works or none work, as miniLock deliberately provides no indication of
// intended recipients.
func DecryptDecryptInfo(diEnc, nonce []byte, ephemPubkey, recipientKey *taber.Keys) (*DecryptInfoEntry, error) {
	plain, err := recipientKey.Decrypt(diEnc, nonce, ephemPubkey)
	if err != nil {
		return nil, ErrCannotDecrypt
	}
	di := new(DecryptInfoEntry)
	err = json.Unmarshal(plain, di)
	if err != nil {
		return nil, err
	}
	return di, nil
}

// ExtractFileInfo pulls out the fileInfo object from the decryptInfo object,
// authenticating encryption from the sender.
func (di *DecryptInfoEntry) ExtractFileInfo(nonce []byte, recipientKey *taber.Keys) (*FileInfo, error) {
	// Return on failure: minilockutils.DecryptionError
	senderPubkey, err := di.SenderPubkey()
	if err != nil {
		return nil, err
	}
	plain, err := recipientKey.Decrypt(di.FileInfoEnc, nonce, senderPubkey)
	if err != nil {
		return nil, ErrCannotDecrypt
	}
	fi := new(FileInfo)
	err = json.Unmarshal(plain, fi)
	if err != nil {
		return nil, err
	}
	return fi, nil
}

// ExtractDecryptInfo iterates through the header using recipientKey and
// attempts to decrypt any DecryptInfoEntry using the provided ephemeral key.
// If unsuccessful after iterating through all decryptInfo objects, returns ErrCannotDecrypt.
func (hdr *miniLockv1Header) ExtractDecryptInfo(recipientKey *taber.Keys) (nonce []byte, DI *DecryptInfoEntry, err error) {
	var (
		ephemKey *taber.Keys
		encDI    []byte
		nonceS   string
	)
	ephemKey = new(taber.Keys)
	ephemKey.Public = hdr.Ephemeral
	if err != nil {
		return nil, nil, err
	}
	// Look for a DI we can decrypt with recipientKey
	// TODO: Make this concurrent!
	for nonceS, encDI = range hdr.DecryptInfo {
		nonce, err := base64.StdEncoding.DecodeString(nonceS)
		if err != nil {
			return nil, nil, err
		}
		DI, err = DecryptDecryptInfo(encDI, nonce, ephemKey, recipientKey)
		if err == ErrCannotDecrypt {
			continue
		} else if err != nil {
			return nil, nil, err
		}
		recipID, err := recipientKey.EncodeID()
		if err != nil {
			return nil, nil, err
		}
		if DI.RecipientID != recipID {
			return nil, nil, ErrBadRecipient
		}
		return nonce, DI, nil
	}
	return nil, nil, ErrCannotDecrypt
}

// ExtractFileInfo tries to pull out a fileInfo all-at-once using a recipientKey.
// It can fail for all the usual reasons including, simply, that the file was not
// encrypted to this recipientKey.
func (hdr *miniLockv1Header) ExtractFileInfo(recipientKey *taber.Keys) (fileinfo *FileInfo, senderID string, err error) {
	nonce, DI, err := hdr.ExtractDecryptInfo(recipientKey)
	if err != nil {
		return nil, "", err
	}
	fileinfo, err = DI.ExtractFileInfo(nonce, recipientKey)
	if err != nil {
		return nil, "", err
	}
	return fileinfo, DI.SenderID, nil
}

// DecryptContents uses a miniLock file's header to attempt decryption of its ciphertext
// all-at-once, enclosing the lower-level operations entirely. It can fail for all
// the usual reasons including that the file simply isn't encrypted to this recipient.
func (hdr *miniLockv1Header) DecryptContents(ciphertext []byte, recipientKey *taber.Keys) (senderID, filename string, contents []byte, err error) {
	var FI *FileInfo
	FI, senderID, err = hdr.ExtractFileInfo(recipientKey)
	if err != nil {
		return "", "", nil, err
	}
	filename, contents, err = FI.DecryptFile(ciphertext)
	if err != nil {
		return "", "", nil, err
	}
	return senderID, filename, contents, nil
}
