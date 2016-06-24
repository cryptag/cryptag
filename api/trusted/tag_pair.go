// Steven Phillips / elimisteve
// 2016.06.22

package trusted

import (
	"github.com/elimisteve/cryptag/types"
)

type TagPair struct {
	Random string `json:"random"`
	Plain  string `json:"plain"`
}

type TagPairs []*TagPair

func FromTagPairs(pairs types.TagPairs) TagPairs {
	out := make(TagPairs, 0, len(pairs))
	for _, pair := range pairs {
		out = append(out, FromTagPair(pair))
	}
	return out
}

func FromTagPair(pair *types.TagPair) *TagPair {
	return &TagPair{Random: pair.Random, Plain: pair.Plain()}
}
