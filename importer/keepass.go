// Steven Phillips / elimisteve
// 2016.06.07

package importer

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/cryptag/cryptag/types"
)

func KeePassCSV(filename string, plaintags []string) (types.Rows, error) {
	csvBody, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(csvBody))

	var rows types.Rows

	for i := 0; ; i++ {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error reading CSV row: %v", err)
		}

		if i == 0 {
			if !eq(line, keePassFormat) {
				return nil, fmt.Errorf("Wrong CSV format; wanted `%s`, got `%s",
					keePassFormatStr, strings.Join(line, ", "))
			}
			// Don't make Row out of column names!
			continue
		}

		row, err := parseKeePassCSVLine(line, plaintags)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func eq(s, t []string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != t[i] {
			return false
		}
	}
	return true
}

var (
	keePassFormat = []string{
		"Group", "Title", "Username", "Password", "URL", "Notes",
	}
	keePassFormatStr = strings.Join(keePassFormat, ", ")
)

func parseKeePassCSVLine(line []string, plaintags []string) (*types.Row, error) {
	if len(line) != 6 {
		return nil, fmt.Errorf("Error parsing CSV line; should have"+
			" 6 parts (%s), has %d", keePassFormatStr, len(line))
	}

	password := []byte(line[3])

	plain := append(plaintags,
		"group:"+line[0],
		"title:"+line[1],
		"login:"+line[2],
		"username:"+line[2],
		"url:"+line[4],
		"notes:"+line[5],

		"type:text",
		"type:password",
	)

	return types.NewRow(password, plain)
}
