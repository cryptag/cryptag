// Steven Phillips / elimisteve
// 2016.06.05

package rowutil

import (
	"testing"

	"github.com/cryptag/cryptag/types"
	"github.com/stretchr/testify/assert"
)

func TestByTagPrefix(t *testing.T) {
	r1, _ := types.NewRowSimple(nil, []string{"tag:01"})
	r2, _ := types.NewRowSimple(nil, []string{"tag:02"})
	r3, _ := types.NewRowSimple(nil, []string{"tag:03"})
	r4, _ := types.NewRowSimple(nil, []string{"tag:04"})

	rows := types.Rows{r2, r1, r4, r3}
	ascend := true
	descend := false

	//
	// Test ascend sort
	//

	rows.Sort(ByTagPrefix("tag:", ascend))

	ascending := types.Rows{r1, r2, r3, r4}

	for i := 0; i < len(rows); i++ {
		assert.Equal(t, rows[i], ascending[i])
	}

	t.Logf("ByTagPrefix's ascend sort finished on all %d rows", len(rows))

	//
	// Test descend sort
	//

	rows.Sort(ByTagPrefix("tag:", descend))

	descending := types.Rows{r4, r3, r2, r1}

	for i := 0; i < len(rows); i++ {
		assert.Equal(t, rows[i], descending[i])
	}

	t.Logf("ByTagPrefix's descend sort finished on all %d rows", len(rows))
}

func TestByTagPrefix2(t *testing.T) {
	r1, _ := types.NewRowSimple(nil, []string{"created:20160605100227-1"})
	r2, _ := types.NewRowSimple(nil, []string{"created:20160605100238-2"})
	r3, _ := types.NewRowSimple(nil, []string{"created:20160605132634-3"})
	r4, _ := types.NewRowSimple(nil, []string{"created:20160605134527-4"})

	rows := types.Rows{r2, r1, r4, r3}
	ascend := true
	descend := false

	//
	// Test ascend sort
	//

	rows.Sort(ByTagPrefix("created:", ascend))

	ascending := types.Rows{r1, r2, r3, r4}

	for i := 0; i < len(rows); i++ {
		assert.Equal(t, rows[i], ascending[i])
	}

	t.Logf("ByTagPrefix's ascend sort done")

	//
	// Test descend sort
	//

	rows.Sort(ByTagPrefix("created:", descend))

	descending := types.Rows{r4, r3, r2, r1}

	for i := 0; i < len(rows); i++ {
		assert.Equal(t, descending[i], rows[i])
	}

	t.Logf("ByTagPrefix's descend sort done")
}
