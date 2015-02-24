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

// Sets each .decrypted and .plainTags field
func (rows Rows) setUnexported() error {
	var err error

	if err = rows.decryptData(); err != nil {
		return err
	}

	if err = rows.setPlainTags(); err != nil {
		return err
	}

	return nil
}

func (rows Rows) decryptData() error {
	var err error
	for i := range rows {
		if err = rows[i].decryptData(); err != nil {
			return fmt.Errorf("Error decrypting data of row %d: %v", i, err)
		}
	}
	return nil
}

func (rows Rows) setPlainTags() error {
	var err error
	for i := range rows {
		if err = rows[i].setPlainTags(); err != nil {
			return fmt.Errorf("Error decrypting tags of row %d: %v", i, err)
		}
	}
	return nil
}
