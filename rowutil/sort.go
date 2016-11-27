package rowutil

import "github.com/cryptag/cryptag/types"

func ByTagPrefix(tagPrefix string, ascending bool) types.RowSorter {
	return func(r1, r2 *types.Row) bool {
		r1Str := TagWithPrefix(r1, tagPrefix)
		r2Str := TagWithPrefix(r2, tagPrefix)
		minLen := min(len(r1Str), len(r2Str))

		for i := 0; i < minLen; i++ {
			if r1Str[i] == r2Str[i] {
				continue
			}
			if r1Str[i] < r2Str[i] {
				// "Smaller" row occurs before if ascending
				return ascending
			}
			if r1Str[i] > r2Str[i] {
				// ...after if descending
				return !ascending
			}
		}
		return ascending
	}
}

func min(n, m int) int {
	if n < m {
		return n
	}
	return m
}
