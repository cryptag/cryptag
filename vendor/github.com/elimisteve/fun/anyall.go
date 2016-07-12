// Steve Phillips / elimisteve
// 2012.08.16

package fun

import (
	"strings"
)

func HasAnyPrefixes(body string, prefixes ...string) bool {
	for _, pre := range prefixes {
		if strings.HasPrefix(body, pre) {
			return true
		}
	}
	return false
}
