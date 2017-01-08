// Steve Phillips / elimisteve
// 2017.01.07

package rowutil_test

import (
	"testing"

	"github.com/cryptag/cryptag/rowutil"
	"github.com/cryptag/cryptag/types"
	"github.com/stretchr/testify/assert"
)

type rowTest struct {
	in  types.Rows
	out []types.Rows
}

func TestToVersionedRows(t *testing.T) {
	r0, _ := types.NewRowSimple(nil, []string{"id:00", "created:0000"})
	r1, _ := types.NewRowSimple(nil, []string{"id:01", "created:0001", "origversionrow:id:00"})
	r2, _ := types.NewRowSimple(nil, []string{"id:02", "created:0002"})
	r3, _ := types.NewRowSimple(nil, []string{"id:03", "created:0003", "origversionrow:id:00"})
	r4, _ := types.NewRowSimple(nil, []string{"id:04", "created:0004", "origversionrow:id:00"})
	r5, _ := types.NewRowSimple(nil, []string{"id:05", "created:0005", "origversionrow:id:DELETED"})
	r6, _ := types.NewRowSimple(nil, []string{"id:06", "created:0006"})
	r7, _ := types.NewRowSimple(nil, []string{"id:07", "created:0007", "origversionrow:id:06"})

	// Ascending

	orig0 := types.Rows{r0, r1, r2, r3, r4, r5, r6, r7}
	want0 := []types.Rows{
		types.Rows{r2},
		types.Rows{r0, r1, r3, r4},
		types.Rows{r5},
		types.Rows{r6, r7},
	}

	orig1 := types.Rows{r1, r3, r0, r4, r7, r5, r2, r6}
	want1 := want0

	ascTests := []rowTest{
		{orig0, want0},
		{orig1, want1},
	}

	ascending := true
	descending := false

	t.Logf("  ** Ascending **\n")
	runTests(t, ascTests, ascending)

	// Descending

	want0 = []types.Rows{
		types.Rows{r6, r7},
		types.Rows{r5},
		types.Rows{r0, r1, r3, r4},
		types.Rows{r2},
	}
	want1 = want0

	descTests := []rowTest{
		{orig0, want0},
		{orig1, want1},
	}

	t.Logf("  ** Descending **\n")
	runTests(t, descTests, descending)
}

func runTests(t *testing.T, tests []rowTest, ascOrDesc bool) {
	for _, tt := range tests {
		got := rowutil.ToVersionedRows(tt.in,
			rowutil.ByTagPrefix("created:", ascOrDesc))
		assert.Equal(t, got, tt.out)
		for _, rows := range got {
			for _, r := range rows {
				t.Logf("%#v\n", r.PlainTags())
			}
			t.Logf("\n")
		}
		t.Logf("\n")
	}
}
