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

func (pairs TagPairs) AllRandom() []string {
	random := make([]string, 0, len(pairs))
	for _, p := range pairs {
		random = append(random, p.Random)
	}
	return random
}

func (pairs TagPairs) HaveAllPlainTags(plaintags []string) (TagPairs, error) {
	var matches TagPairs
	for _, plain := range plaintags {
		for _, pair := range pairs {
			if pair.plain == plain {
				matches = append(matches, pair)
				break
			}
		}
	}
	if len(matches) == 0 {
		return nil, ErrTagPairNotFound
	}
	return matches, nil
}

func (pairs TagPairs) HaveAllRandomTags(randomtags []string) (TagPairs, error) {
	var matches TagPairs
	for _, random := range randomtags {
		for _, pair := range pairs {
			if pair.Random == random {
				matches = append(matches, pair)
				break
			}
		}
	}
	if len(matches) == 0 {
		return nil, ErrTagPairNotFound
	}
	return matches, nil
}
