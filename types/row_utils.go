// Steve Phillips / elimisteve
// 2015.02.24

package types

import "github.com/elimisteve/fun"

// func ParseRows(b []byte) (Rows, error) {
// 	var rows Rows
// 	var err error

// 	if err = json.Unmarshal(b, &rows); err != nil {
// 		return nil, err
// 	}

// 	if err = rows.setUnexported(); err != nil {
// 		return nil, err
// 	}

// 	return rows, nil
// }

func GetRowsFrom(url string) (Rows, error) {
	var rows Rows
	var err error

	if err = fun.FetchInto(url, HttpGetTimeout, &rows); err != nil {
		return nil, err
	}

	if err = rows.setUnexported(); err != nil {
		return nil, err
	}

	return rows, nil
}
