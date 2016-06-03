package rowutil

import (
	"strings"

	"github.com/elimisteve/cryptag/types"
)

// TagWithPrefix iterates over r's plaintags to grab the data stored
// in said tags, preferring the prefixes in the order they are passed
// in.
//
// Example use: TagWithPrefix(r, "filename:", "id:"), which will try
// to return r's filename, then try to return its ID, and if both
// fail, will return empty string.
func TagWithPrefix(r *types.Row, prefixes ...string) string {
	for _, prefix := range prefixes {
		for _, tag := range r.PlainTags() {
			if strings.HasPrefix(tag, prefix) {
				return strings.TrimPrefix(tag, prefix)
			}
		}
	}
	return ""
}
