package minilock

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
)

func randBytes(i int) ([]byte, error) {
	randBytes := make([]byte, i)
	read, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}
	if read != i {
		return nil, ErrInsufficientEntropy
	}
	return randBytes, nil
}

func makeFullNonce() ([]byte, error) {
	return randBytes(24)
}

func toLittleEndian(i int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, i)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func fromLittleEndian(buf []byte) (int32, error) {
	var output int32
	bufReader := bytes.NewReader(buf)
	err := binary.Read(bufReader, binary.LittleEndian, &output)
	if err != nil {
		return 0, err
	}
	return output, nil
}
