// Steve Phillips / elimisteve
// 2015.02.24

package types

import "fmt"

type Rows []*Row

func (rows Rows) String() string {
	var s string
	for _, row := range rows {
		s += fmt.Sprintf("%#v\n", row)
	}
	return s
}

func (rows Rows) FilterByRandomTags(random []string) (matches Rows) {
	matches = rows[:] // TODO: Ensure this is good enough
	for _, randtag := range random {
		matches = matches.FilterByRandomTag(randtag)
	}
	return matches
}

func (rows Rows) FilterByRandomTag(randtag string) (matches Rows) {
	for _, row := range rows {
		if row.HasRandomTag(randtag) {
			matches = append(matches, row)
		}
	}
	return matches
}

func (rows Rows) ExcludeByRandomTags(random []string) (nonmatches Rows) {
	nonmatches = rows[:]
	for _, randtag := range random {
		nonmatches = nonmatches.ExcludeByRandomTag(randtag)
	}
	return nonmatches
}

func (rows Rows) ExcludeByRandomTag(randtag string) (nonmatches Rows) {
	for _, row := range rows {
		if !row.HasRandomTag(randtag) {
			nonmatches = append(nonmatches, row)
		}
	}
	return nonmatches
}

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
