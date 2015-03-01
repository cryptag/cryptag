// Steve Phillips / elimisteve
// 2015.02.24

package types

import "fmt"

type TagPairs []*TagPair

func (pairs TagPairs) String() string {
	var s string
	for _, pair := range pairs {
		s += fmt.Sprintf("%#v\n", pair)
	}
	return s
}

func (pairs TagPairs) AllPlain() []string {
	plain := make([]string, 0, len(pairs))
	for _, p := range pairs {
		plain = append(plain, p.plain)
	}
	return plain
}

func (pairs TagPairs) HaveAllPlainTags(plaintags []string) TagPairs {
	var matches TagPairs
	for _, pair := range pairs {
		for _, plain := range plaintags {
			if pair.plain == plain {
				matches = append(matches, pair)
				break
			}
		}
	}
	return matches
}
