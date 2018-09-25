// Steve Phillips / elimisteve
// 2018.09.24

package exporter

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cryptag/cryptag/backend"
	"github.com/cryptag/cryptag/rowutil"
	"github.com/cryptag/cryptag/types"
)

func ToLastPassCSV(bk backend.Backend, filename string, plaintags []string) error {
	if filename == "" {
		return errors.New("filename cannot be empty!")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error creating file '%s': %v", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	rows, err := backend.RowsFromPlainTags(bk, nil, plaintags)
	if err != nil {
		return fmt.Errorf("Error getting rows with plaintags `%#v`: %v",
			plaintags, err)
	}

	err = writer.Write(lastpassFormat)
	if err != nil {
		return fmt.Errorf("Error writing header line `%#v` to file `%s`: %v",
			lastpassFormat, filename, err)
	}

	for _, row := range rows {
		line, err := createLastpassLine(row)
		if err != nil {
			return fmt.Errorf("Error creating line in LastPass CSV file: %v", err)
		}
		if err = writer.Write(line); err != nil {
			return fmt.Errorf("Error writing line '%#v' to %s: %v",
				line, filename, err)
		}
	}

	return nil
}

var (
	lastpassFormat = []string{
		"url", "type", "username", "password", "hostname", "extra", "name", "grouping",
	}
)

func createLastpassLine(row *types.Row) (line []string, err error) {
	pwType := "password"
	types := rowutil.TagsWithPrefix(row, "type:")
	for _, typ := range types {
		if typ != "type:text" && typ != "type:file" && typ != "type:password" {
			pwType = typ[5:]
			break
		}
	}

	url := rowutil.TagWithPrefixStripped(row, "url:", "domain:", "site:", "website:")
	username := rowutil.TagWithPrefixStripped(row, "login:", "username:", "email:")
	password := string(row.Decrypted())
	hostname := rowutil.TagWithPrefixStripped(row, "domain:", "hostname:", "url:", "site:", "website:")
	extra := strings.Join(row.PlainTags(), ", ")
	name := rowutil.TagWithPrefixStripped(row, "name:")
	grouping := rowutil.TagWithPrefixStripped(row, "grouping:")

	return []string{
		url,
		pwType,
		username,
		password,
		hostname,
		extra,
		name,
		grouping,
	}, nil
}
