package rowutil

import (
	"strings"

	"github.com/cryptag/cryptag/types"
)

// TagWithPrefixStripped does the same thing as TagWithPrefix except
// it also strips off the prefix from the tags found with said
// prefix. (Ya get me?)
func TagWithPrefixStripped(r *types.Row, prefixes ...string) string {
	stripped := true
	return tagWithPrefix(r, prefixes, stripped)
}

// TagWithPrefix iterates over r's plaintags to grab the data stored
// in said tags, preferring the prefixes in the order they are passed
// in.
//
// Example use: TagWithPrefix(r, "filename:", "id:"), which will try
// to return r's filename, then try to return its ID, and if both
// fail, will return empty string.
func TagWithPrefix(r *types.Row, prefixes ...string) string {
	stripped := false
	return tagWithPrefix(r, prefixes, stripped)
}

func tagWithPrefix(r *types.Row, prefixes []string, stripped bool) string {
	for _, prefix := range prefixes {
		for _, tag := range r.PlainTags() {
			if strings.HasPrefix(tag, prefix) {
				if stripped {
					return strings.TrimPrefix(tag, prefix)
				}
				return tag
			}
		}
	}
	return ""
}

// TagsWithPrefixStripped does the same thing as TagsWithPrefix,
// except it also strips off the prefix from each of the PlainTags
// found with said prefix.
func TagsWithPrefixStripped(r *types.Row, prefix string) []string {
	stripped := true
	return tagsWithPrefix(r, prefix, stripped)
}

// TagsWithPrefix returns r's PlainTags that begin with prefix.
func TagsWithPrefix(r *types.Row, prefix string) []string {
	stripped := false
	return tagsWithPrefix(r, prefix, stripped)
}

func tagsWithPrefix(r *types.Row, prefix string, stripped bool) []string {
	var plaintags []string

	for _, tag := range r.PlainTags() {
		if strings.HasPrefix(tag, prefix) {
			if stripped {
				plaintags = append(plaintags, strings.TrimPrefix(tag, prefix))
				continue
			}
			plaintags = append(plaintags, tag)
		}
	}

	return plaintags
}
