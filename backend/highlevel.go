// Steven Phillips / elimisteve
// 2016.03.28

package backend

import (
	"log"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/types"
)

func RowsFromPlainTags(bk Backend, pairs types.TagPairs, plaintags cryptag.PlainTags) (types.Rows, error) {
	return getRows(bk, pairs, plaintags, bk.RowsFromRandomTags)
}

func ListRowsFromPlainTags(bk Backend, pairs types.TagPairs, plaintags cryptag.PlainTags) (types.Rows, error) {
	return getRows(bk, pairs, plaintags, bk.ListRows)
}

func getRows(bk Backend, pairs types.TagPairs, plaintags cryptag.PlainTags, fetchByRandom func(cryptag.RandomTags) (types.Rows, error)) (types.Rows, error) {
	if pairs == nil {
		var err error
		pairs, err = bk.AllTagPairs()
		if err != nil {
			return nil, err
		}
	}

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

func DeleteRows(bk Backend, pairs types.TagPairs, plaintags cryptag.PlainTags) error {
	if pairs == nil {
		var err error
		pairs, err = bk.AllTagPairs()
		if err != nil {
			return err
		}
	}

	matches, err := pairs.WithAllPlainTags(plaintags)
	if err != nil {
		return err
	}

	randtags := matches.AllRandom()

	if types.Debug {
		log.Printf("Deleting rows with PlainTags `%v` / RandomTags `%v`\n",
			plaintags, randtags)
	}

	return bk.DeleteRows(randtags)
}

func CreateRow(bk Backend, pairs types.TagPairs, rowData []byte, plaintags []string) (*types.Row, error) {
	if types.Debug {
		log.Printf("Creating row with data of length %d and tags `%#v`\n",
			len(rowData), plaintags)
	}

	row, err := types.NewRow(rowData, plaintags)
	if err != nil {
		return nil, err
	}

	if pairs == nil {
		pairs, err = bk.AllTagPairs()
		if err != nil {
			return nil, err
		}
	}

	err = PopulateRowBeforeSave(bk, row, pairs)
	if err != nil {
		return nil, err
	}

	err = bk.SaveRow(row)
	if err != nil {
		return nil, err
	}

	return row, nil
}
