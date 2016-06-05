// Steve Phillips / elimisteve
// 2015.02.24

package types

import (
	"fmt"
	"log"
	"sort"
)

type Rows []*Row

func (rows Rows) String() string {
	var s string
	for _, row := range rows {
		s += fmt.Sprintf("%#v\n", row)
	}
	return s
}

func (rows Rows) Format() string {
	var s string
	for _, row := range rows {
		s += "\n" + row.Format()
	}
	return s
}

// WithAllRandomTags returns the Rows within rows that has all the
// random strings in random
func (rows Rows) WithAllRandomTags(random []string) Rows {
	// Copy rows
	matches := make(Rows, len(rows))
	copy(matches, rows)

	// If any row doesn't have any of the random tags, remove it
	for _, randtag := range random {
		for i := range matches {
			if !matches[i].HasRandomTag(randtag) {
				log.Printf("%+v's random tags don't include `%s`\n",
					matches[i], randtag)
				matches = append(matches[:i], matches[i+1:]...)
				break
			}
		}
	}
	return matches
}

func (rows Rows) Populate(key *[32]byte, pairs TagPairs) error {
	// TODO: Benchmark whether parallelizing would increase
	// performance
	for i := range rows {
		if err := rows[i].Populate(key, pairs); err != nil {
			return err
		}
	}
	return nil
}

func (rows Rows) Sort(less func(r1, r2 *Row) bool) {
	rs := rowSorter{rows, less}
	sort.Sort(rs)
}

type rowSorter struct {
	rows Rows
	less func(r1, r2 *Row) bool
}

func (rs rowSorter) Swap(i, j int) {
	rs.rows[i], rs.rows[j] = rs.rows[j], rs.rows[i]
}

func (rs rowSorter) Len() int { return len(rs.rows) }

func (rs rowSorter) Less(i, j int) bool {
	return rs.less(rs.rows[i], rs.rows[j])
}
