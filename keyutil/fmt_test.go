// Steve Phillips / elimisteve
// 2017.01.05

package keyutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type keyTest struct {
	key    *[32]byte
	keystr string
}

var keyTests = []keyTest{
	{
		&[32]byte{},
		"0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0",
	},
	{
		&[32]byte{
			0, 1, 2, 3, 4, 5, 6, 7,
			0, 1, 2, 3, 4, 5, 6, 7,
			0, 1, 2, 3, 4, 5, 6, 7,
			0, 1, 2, 3, 4, 5, 6, 7,
		},
		"0,1,2,3,4,5,6,7,0,1,2,3,4,5,6,7,0,1,2,3,4,5,6,7,0,1,2,3,4,5,6,7",
	},
}

func TestFormat(t *testing.T) {
	for _, test := range keyTests {
		got := Format(test.key)
		assert.Equal(t, got, test.keystr)
	}
}

var keyTestsErr = []keyTest{
	{
		nil,
		"<nil>",
	},
}
