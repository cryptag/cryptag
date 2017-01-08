package rowutil

import (
	"log"
	"strings"

	"github.com/cryptag/cryptag/types"
)

// ToVersionedRows copies origRows then segments these rows by
// versioned row using the origversionrow:... prefix convention. The
// returned 2D slice is a slice of slices of versions, where each
// inner slice and the outer one are sorted using rowLess.
//
// origRows's ordering is not affected.
func ToVersionedRows(origRows types.Rows, rowLess func(r1, r2 *types.Row) bool) []types.Rows {
	// Don't change the ordering of origRows; make a copy
	rows := make(types.Rows, len(origRows))
	copy(rows, origRows)

	// Every inner slice will remain in chronological order (ascending)
	rows.Sort(ByTagPrefix("created:", true))

	// 2D return value
	var rrows []types.Rows

	ORIG_VERSION_ROW_PREFIX := "origversionrow:"

	// Create map[groupTag]Rows to group the rows together by the
	// ID-tag of the original Row that has since been versioned
	mRows := make(map[string]types.Rows, len(rows))
	for _, r := range rows {
		tag := TagWithPrefix(r, ORIG_VERSION_ROW_PREFIX, "id:")
		if tag == "" {
			log.Printf("Row with tags %#v has no ID-tag!\n", r.PlainTags())
			continue
		}
		if strings.HasPrefix(tag, ORIG_VERSION_ROW_PREFIX) {
			// origversionrow:id:... -> id:...
			tag = tag[len(ORIG_VERSION_ROW_PREFIX):]
		}
		// assert: tag is of the form id:..., where this tag is the
		// ID-tag of the original version of every Row in rows
		mRows[tag] = append(mRows[tag], r)
	}

	// Each version slice is now individually ordered, but we need to
	// return them in a [][]*Row. What should the ordering of this
	// slice of slices be? Answer: ordered by the last member of each
	// version slice (in either ascending or descending order, as
	// determined by rowLess).

	lasts := make(types.Rows, 0, len(mRows))
	lastToRows := make(map[*types.Row]types.Rows, len(lasts))

	for _, rowGroup := range mRows {
		last := rowGroup[len(rowGroup)-1]
		lasts = append(lasts, last)
		// Map the last element of each slice to its containing slice
		lastToRows[last] = rowGroup
	}
	lasts.Sort(rowLess)

	for _, fr := range lasts {
		rrows = append(rrows, lastToRows[fr])
	}

	return rrows
}
