// Steven Phillips / elimisteve
// 2016.03.28

package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/elimisteve/cryptag"
	"github.com/elimisteve/cryptag/keyutil"
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

func CreateFileRow(bk Backend, pairs types.TagPairs, filename string, plaintags []string) (*types.Row, error) {
	rowData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file `%s`: %v\n", filename, err)
	}

	plaintags = append(plaintags, "type:file", "filename:"+filepath.Base(filename))

	return CreateRow(bk, pairs, rowData, plaintags)
}

// newKey can be of type *[32]byte, []byte (with length 32), or a
// string to be parsed with keyutil.Parse.
func UpdateKey(bk Backend, newKey interface{}) error {
	var goodKey *[32]byte

	switch newKey := newKey.(type) {
	case *[32]byte:
		goodKey = newKey
	case []byte:
		k, err := cryptag.ConvertKey(newKey)
		if err != nil {
			return err
		}
		goodKey = k
	case string:
		k, err := keyutil.Parse(newKey)
		if err != nil {
			return err
		}
		goodKey = k
	default:
		panic(fmt.Sprintf("Key of invalid type: %T", newKey))
	}

	if goodKey == nil {
		return fmt.Errorf("New key %v (of type %[1]T) passed in,"+
			" yet goodKey not set correctly", newKey)
	}

	cfg, err := bk.ToConfig()
	if err != nil {
		return err
	}

	cfg.Key = goodKey

	return cfg.Update(cryptag.BackendPath)
}
