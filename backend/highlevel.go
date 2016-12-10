// Steven Phillips / elimisteve
// 2016.03.28

package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/cryptag/cryptag"
	"github.com/cryptag/cryptag/keyutil"
	"github.com/cryptag/cryptag/rowutil"
	"github.com/cryptag/cryptag/types"
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
		pairs, err = bk.AllTagPairs(nil)
		if err != nil {
			return nil, err
		}
	}

	if len(pairs) == 0 {
		return nil, types.ErrTagPairNotFound
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
		pairs, err = bk.AllTagPairs(nil)
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
		pairs, err = bk.AllTagPairs(nil)
		if err != nil {
			return nil, err
		}
	}

	_, err = PopulateRowBeforeSave(bk, row, pairs)
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
		return nil, fmt.Errorf("Error reading file `%s`: %v\n", filename, err)
	}

	plaintags = append(plaintags, "type:file", "filename:"+filepath.Base(filename))

	// Add tag based on filetype (e.g., type:pdf)
	fileExt := getFileExt(filename)
	if fileExt != "" {
		plaintags = append(plaintags, "type:"+fileExt)
	}

	return CreateRow(bk, pairs, rowData, plaintags)
}

func CreateJSONRow(bk Backend, pairs types.TagPairs, obj interface{}, plaintags []string) (*types.Row, error) {
	rowData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return CreateRow(bk, pairs, rowData, plaintags)
}

// UpdateRow creates a new version of the Row whose ID tag is
// prevIDTag, but replaces the "id:..." and "created:..." tags. It
// then adds an "origversionrow:..." tag if one does not already exist
// which points to the ID tag of the original Row being versioned
// here.
func UpdateRow(bk Backend, pairs types.TagPairs, prevIDTag string, newData []byte) (*types.Row, error) {
	var err error
	if pairs == nil {
		pairs, err = bk.AllTagPairs(nil)
		if err != nil {
			return nil, err
		}
	}

	oldRows, err := ListRowsFromPlainTags(bk, pairs, []string{prevIDTag})
	if err != nil {
		return nil, err
	}

	if len(oldRows) != 1 {
		return nil, fmt.Errorf("Query tag `%s` returned %d Rows, not 1",
			prevIDTag, len(oldRows))
	}

	oldRow := oldRows[0]

	return UpdateRowAdvanced(bk, pairs, oldRow, newData, oldRow.PlainTags())
}

// UpdateRowAdvanced creates a new version of oldRow but with updated
// tags that begin with newishTags, remove the "id:...",
// "created:...", and "all" tags, and will add an "origversionrow:..."
// tag that points to the ID tag of the original Row being versioned
// here (or keep the existing "origversionrow:..." tag).
//
// If the plaintags that the new, updated row should have doesn't
// require any pre-processing, newishTags can simply be
// oldRow.PlainTags().  (You may want your pre-processing step to add
// tags like `prevversionrow:...` or user-specified tags.)
func UpdateRowAdvanced(bk Backend, pairs types.TagPairs, oldRow *types.Row, newData []byte, newishTags []string) (*types.Row, error) {
	var origIDTag string

	var newTags []string
	for i := range newishTags {
		if strings.HasPrefix(newishTags[i], "id:") {
			// Added by CreateRow
			continue
		}
		if strings.HasPrefix(newishTags[i], "created:") {
			// Added by CreateRow
			continue
		}
		if newishTags[i] == "all" {
			// Added by CreateRow
			continue
		}

		if strings.HasPrefix(newishTags[i], "origversionrow:") {
			// Re-added below
			origIDTag = newishTags[i]
			continue
		}

		newTags = append(newTags, newishTags[i])
	}

	// If oldRow is the original version (if there aren't any previous
	// versions of this Row), set this new row's origversionrow to
	// oldRow
	if origIDTag == "" {
		oldRowIDTag := rowutil.TagWithPrefix(oldRow, "id:")
		newTags = append(newTags, "origversionrow:"+oldRowIDTag)
	} else {
		newTags = append(newTags, origIDTag)
	}

	return CreateRow(bk, pairs, newData, newTags)
}

// UpdateFileRow finds the Row uniquely picked out by prevIDTag then
// creates a new Row consisting of the contents of newFilename and the
// tags from the Row being updated (after replacing id:...,
// created:..., and filename:..., and adding a
// origversionrow:... tag).
func UpdateFileRow(bk Backend, pairs types.TagPairs, prevIDTag string, newFilename string) (*types.Row, error) {
	var err error
	if pairs == nil {
		pairs, err = bk.AllTagPairs(nil)
		if err != nil {
			return nil, err
		}
	}

	rows, err := ListRowsFromPlainTags(bk, pairs, []string{prevIDTag})
	if err != nil {
		return nil, err
	}

	// TODO: Could create UpdateLastFileRow variant of this func that
	// instead lets the query return >1 Row, and it uses the last one
	// (sorted by `created:...`) as the previous version.
	if len(rows) != 1 {
		return nil, fmt.Errorf("Query tag `%v` must uniquely specify 1 Row, not %d",
			prevIDTag, len(rows))
	}

	oldRow := rows[0]
	oldTags := oldRow.PlainTags()
	oldFilename := rowutil.TagWithPrefixStripped(oldRow, "filename:")

	// Sure this a file?
	if !oldRow.HasPlainTag("type:file") {
		return nil, fmt.Errorf(`Row %s is not a file (no "type:file" tag)`,
			rowutil.TagWithPrefix(oldRow, "id:"))
	}

	// Determine which old `type:${file_extension}` tag to remove below

	// Could be "" if no file ext exists
	oldFileExt := getFileExt(oldFilename)

	// Copy all oldTags except "filename:...", "type:"+oldFileExt. Not
	// removing "id...", "created:...", and "all" because UpdateRow
	// will do that.

	var newTags []string
	for i := range oldTags {
		if strings.HasPrefix(oldTags[i], "filename:") {
			// New filename added below
			continue
		}
		if oldTags[i] == "type:"+oldFileExt {
			// New `type:...` tag added below for the new file ext
			continue
		}

		newTags = append(newTags, oldTags[i])
	}

	// Add "filename:..." and "type:"+newFileExt tags
	newTags = append(newTags, "filename:"+filepath.Base(newFilename))

	if fileExt := getFileExt(newFilename); fileExt != "" {
		newTags = append(newTags, "type:"+fileExt)
	}

	// Read file data

	// TODO: Need smarter file streaming logic...
	newData, err := ioutil.ReadFile(newFilename)
	if err != nil {
		return nil, err
	}

	return UpdateRowAdvanced(bk, pairs, oldRow, newData, newTags)
}

func getFileExt(filenameOrPath string) string {
	var fileExt string

	lastDot := strings.LastIndex(filenameOrPath, ".")
	if lastDot != -1 {
		fileExt = filenameOrPath[lastDot+1:]
	}

	return strings.ToLower(fileExt)
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
