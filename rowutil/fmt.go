// Steven Phillips / elimisteve
// 2016.05.11

package rowutil

import (
	"github.com/cryptag/cryptag/types"
)

// MapToStrings applies mapper to every *Row in rows
func MapToStrings(mapper func(*types.Row) string, rows types.Rows) []string {
	rowStrs := make([]string, len(rows))
	for i := range rows {
		rowStrs[i] = mapper(rows[i])
	}
	return rowStrs
}
