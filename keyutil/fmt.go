// Steven Phillips / elimisteve
// 2016.06.05

package keyutil

import (
	"fmt"
)

func Format(key *[32]byte) string {
	if key == nil {
		return "<nil>"
	}
	k := *key

	return FormatSlice(k[:])
}

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
