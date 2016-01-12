// Steven Phillips / elimisteve
// 2016.01.11

package types

import (
	"strings"
)

func RowTagWithPrefix(r *Row, prefix string) string {
	for _, t := range r.PlainTags() {
		if strings.HasPrefix(t, prefix) {
			return strings.TrimPrefix(t, prefix)
		}
	}
	return ""
}
