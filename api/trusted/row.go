// Steven Phillips / elimisteve
// 2016.06.22

package trusted

import (
	"github.com/cryptag/cryptag/types"
)

type Row struct {
	Unencrypted []byte   `json:"unencrypted"` // types.Row.Decrypted()
	PlainTags   []string `json:"plaintags"`   // types.Row.PlainTags()
}

type Rows []*Row

func FromRows(rows types.Rows) Rows {
	out := make(Rows, 0, len(rows))
	for _, in := range rows {
		out = append(out, FromRow(in))
	}
	return out
}

func FromRow(row *types.Row) *Row {
	return &Row{Unencrypted: row.Decrypted(), PlainTags: row.PlainTags()}
}

type RowUpdate struct {
	Unencrypted  []byte `json:"unencrypted"` // types.Row.Decrypted()
	OldVersionID string `json:"old_version_id_tag"`
}

func FromRows2D(rrows []types.Rows) []Rows {
	vrrows := make([]Rows, 0, len(rrows))
	for _, rows := range rrows {
		vrrows = append(vrrows, FromRows(rows))
	}
	return vrrows
}
