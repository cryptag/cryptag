// Steve Phillips / elimisteve
// 2015.02.24

package types

import (
	"fmt"
	"log"
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
		s += row.Format()
	}
	return s
}

func (rows Rows) FirstToClipboard() error {
	if len(rows) == 0 {
		return nil // TODO: Is this right?
	}
	return rows[0].ToClipboard()
}

// HaveAllRandomTags returns the Rows within rows that has all the
// random strings in random
func (rows Rows) HaveAllRandomTags(random []string) Rows {
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
