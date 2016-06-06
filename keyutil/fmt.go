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

	kStr := fmt.Sprintf("%d", k[0])

	for i := 1; i < len(k); i++ {
		kStr += fmt.Sprintf(",%d", k[i])
	}
	return kStr
}
