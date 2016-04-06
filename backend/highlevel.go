// Steven Phillips / elimisteve
// 2016.03.28

package backend

import (
	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/types"
)

func RowsFromPlainTags(bk Backend, plaintags cryptag.PlainTags, pairs types.TagPairs) (types.Rows, error) {
	matches, err := pairs.HaveAllPlainTags(plaintags)
	if err != nil {
		return nil, err
	}

	rows, err := bk.RowsFromRandomTags(matches.AllRandom())
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, types.ErrRowsNotFound
	}

	if err := rows.Populate(bk.Key(), pairs); err != nil {
		return nil, err
	}

	return rows, nil
}

func DeleteRows(bk Backend, plaintags cryptag.PlainTags, pairs types.TagPairs) error {
	matches, err := pairs.HaveAllPlainTags(plaintags)
	if err != nil {
		return err
	}

	// Delete rows tagged with the random strings in pairs
	return bk.DeleteRows(matches.AllRandom())
}
