// Steven Phillips / elimisteve
// 2016.06.05

package keyutil

import (
	"fmt"
)

// Format formats the given key in the format a,b,c,...,z which is
// human-readable and suitable to be read by backend.UpdateKey().
func Format(key *[32]byte) string {
	if key == nil {
		return "<nil>"
	}
	k := *key

	return FormatSlice(k[:])
}

// Format formats the given byte slice (probably a nonce slice-ified
// key) in the format a,b,c,...,z which is human-readable and suitable
// to be read by backend.UpdateKey().
func FormatSlice(b []byte) string {
	if b == nil {
		return "<nil>"
	}
	if len(b) == 0 {
		return ""
	}

	bStr := fmt.Sprintf("%d", b[0])

	for i := 1; i < len(b); i++ {
		bStr += fmt.Sprintf(",%d", b[i])
	}
	return bStr
}
