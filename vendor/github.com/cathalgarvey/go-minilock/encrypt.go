package minilock

import (
	"encoding/base64"
	"encoding/json"

	"github.com/cathalgarvey/go-minilock/taber"
	"github.com/dchest/blake2s"
)

// EncryptFileToFileInfo symmetrically encrypts a file or plaintext and returns
// the fileInfo object to decrypt it and the raw ciphertext. This operation is
// technically independent of miniLock and could be used for other crypto-schemes
// as a handy way to encrypt files symmetrically.
func EncryptFileToFileInfo(filename string, filecontents []byte) (FI *FileInfo, ciphertext []byte, err error) {
	var (
		DI *taber.DecryptInfo
	)
	DI, err = taber.NewDecryptInfo()
	if err != nil {
		return nil, nil, err
	}
	return encryptFileToFileInfo(DI, filename, filecontents)
}

// Separated from the above for testing purposes; deterministic ciphertext.
func encryptFileToFileInfo(DI *taber.DecryptInfo, filename string, filecontents []byte) (FI *FileInfo, ciphertext []byte, err error) {
	var hash [32]byte
	if filecontents == nil || len(filecontents) == 0 {
		return nil, nil, ErrNilPlaintext
	}
	ciphertext, err = DI.Encrypt(filename, filecontents)
	if err != nil {
		return nil, nil, err
	}
	hash = blake2s.Sum256(ciphertext)
	FI = new(FileInfo)
	FI.FileKey = DI.Key
	FI.FileNonce = DI.BaseNonce
	FI.FileHash = hash[:]
	return FI, ciphertext, nil
}

// NewDecryptInfoEntry creates a decryptInfo entry for the given fileInfo to the intended recipientKey,
// from senderKey.
func NewDecryptInfoEntry(nonce []byte, fileinfo *FileInfo, senderKey, recipientKey *taber.Keys) (*DecryptInfoEntry, error) {
	encodedFi, err := json.Marshal(fileinfo)
	if err != nil {
		return nil, err
	}
	cipherFi, err := senderKey.Encrypt(encodedFi, nonce, recipientKey)
	if err != nil {
		return nil, err
	}
	senderID, err := senderKey.EncodeID()
	if err != nil {
		return nil, err
	}
	recipientID, err := recipientKey.EncodeID()
	if err != nil {
		return nil, err
	}
	return &DecryptInfoEntry{SenderID: senderID, RecipientID: recipientID, FileInfoEnc: cipherFi}, nil
}

// EncryptDecryptInfo encrypts a decryptInfo struct using the ephemeral pubkey
// and the same nonce as the enclosed fileInfo.
func EncryptDecryptInfo(di *DecryptInfoEntry, nonce []byte, ephemKey, recipientKey *taber.Keys) ([]byte, error) {
	plain, err := json.Marshal(di)
	if err != nil {
		return nil, err
	}
	// NaClKeypair.Encrypt(plaintext, nonce []byte, to *NaClKeypair) (ciphertext []byte, err error)
	diEnc, err := ephemKey.Encrypt(plain, nonce, recipientKey)
	if err != nil {
		return nil, err
	}
	return diEnc, nil
}

func (hdr *miniLockv1Header) addFileInfo(fileInfo *FileInfo, ephem, sender *taber.Keys, recipients ...*taber.Keys) error {
	for _, recipientKey := range recipients {
		nonce, rgerr := makeFullNonce()
		if rgerr != nil {
			return rgerr
		}
		// NewDecryptInfoEntry(nonce []byte, fileinfo *FileInfo, senderKey, recipientKey *taber.Keys) (*DecryptInfoEntry, error) {
		DI, rgerr := NewDecryptInfoEntry(nonce, fileInfo, sender, recipientKey)
		if rgerr != nil {
			return rgerr
		}
		encDI, rgerr := EncryptDecryptInfo(DI, nonce, ephem, recipientKey)
		if rgerr != nil {
			return rgerr
		}
		nonceS := base64.StdEncoding.EncodeToString(nonce)
		hdr.DecryptInfo[nonceS] = encDI
	}
	return nil
}

// EncryptFileContentsWithStrings is an entry point that largely defines "normal"
// miniLock behaviour. If sendToSender is true, then the sender's ID is added to recipients.
func EncryptFileContentsWithStrings(filename string, fileContents []byte, senderEmail, senderPassphrase string, sendToSender bool, recipientIDs ...string) (miniLockContents []byte, err error) {
	var (
		senderKey, thisRecipient *taber.Keys
		recipientKeyList         []*taber.Keys
		thisID                   string
	)
	senderKey, err = taber.FromEmailAndPassphrase(senderEmail, senderPassphrase)
	if err != nil {
		return nil, err
	}
	defer senderKey.Wipe()
	if sendToSender {
		thisID, err = senderKey.EncodeID()
		if err != nil {
			return nil, err
		}
		recipientIDs = append(recipientIDs, thisID)
	}
	recipientKeyList = make([]*taber.Keys, 0, len(recipientIDs))
	// TODO: Randomise iteration here?
	for _, thisID = range recipientIDs {
		thisRecipient, err = taber.FromID(thisID)
		if err != nil {
			return nil, err
		}
		recipientKeyList = append(recipientKeyList, thisRecipient)
	}
	miniLockContents, err = EncryptFileContents(filename, fileContents, senderKey, recipientKeyList...)
	if err != nil {
		return nil, err
	}
	return miniLockContents, nil
}

// EncryptFileContents is an entry point for encrypting byte slices from a prepared
// sender key to prepared recipient keys. EncryptFileContentsWithStrings is much
// easier to use for most applications.
func EncryptFileContents(filename string, fileContents []byte, sender *taber.Keys, recipients ...*taber.Keys) (miniLockContents []byte, err error) {
	var (
		hdr        *miniLockv1Header
		ephem      *taber.Keys
		ciphertext []byte
		fileInfo   *FileInfo
	)
	hdr, ephem, err = prepareNewHeader()
	if err != nil {
		return nil, err
	}
	fileInfo, ciphertext, err = EncryptFileToFileInfo(filename, fileContents)
	if err != nil {
		return nil, err
	}
	err = hdr.addFileInfo(fileInfo, ephem, sender, recipients...)
	if err != nil {
		return nil, err
	}
	miniLockContents = make([]byte, 0, 8+4+hdr.encodedLength()+len(ciphertext))
	miniLockContents, err = hdr.stuffSelf(miniLockContents)
	if err != nil {
		return nil, err
	}
	miniLockContents = append(miniLockContents, ciphertext...)
	return miniLockContents, nil
}
