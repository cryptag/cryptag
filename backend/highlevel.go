// Steven Phillips / elimisteve
// 2016.03.28

package backend

import (
	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/types"
)

func RowsFromPlainTags(bk Backend, plaintags cryptag.PlainTags, pairs types.TagPairs) (types.Rows, error) {
	return getRows(bk, plaintags, pairs, bk.RowsFromRandomTags)
}

func ListRowsFromPlainTags(bk Backend, plaintags cryptag.PlainTags, pairs types.TagPairs) (types.Rows, error) {
	return getRows(bk, plaintags, pairs, bk.ListRows)
}

func getRows(bk Backend, plaintags cryptag.PlainTags, pairs types.TagPairs, fetchByRandom func(cryptag.RandomTags) (types.Rows, error)) (types.Rows, error) {
	matches, err := pairs.WithAllPlainTags(plaintags)
	if err != nil {
		return nil, err
	}

	rows, err := fetchByRandom(matches.AllRandom())
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
	matches, err := pairs.WithAllPlainTags(plaintags)
	if err != nil {
		return err
	}

	// Delete rows tagged with the random strings in pairs
	return bk.DeleteRows(matches.AllRandom())
}
