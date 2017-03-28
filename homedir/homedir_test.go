// Steve Phillips / elimisteve
// 2017.03.27

package homedir

import (
	"path"
	"testing"

	gohomedir "github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
)

var home = ""

func init() {
	dir, err := gohomedir.Dir()
	if err != nil {
		panic(err)
	}
	home = dir
}

func TestCollapse(t *testing.T) {
	type dirs struct {
		path          string
		wantCollapsed string
	}

	tests := []dirs{
		{
			"",
			"",
		},
		{
			"/",
			"/",
		},
		{
			"/a",
			"/a",
		},
		{
			path.Join(home, ".cryptag", "data"),
			path.Join("~", ".cryptag", "data"),
		},
		{
			"/home/notexistentuser!/.cryptag/data",
			"/home/notexistentuser!/.cryptag/data",
		},
		{
			`Z:\Users\notexistentuser\.cryptag\data`,
			`Z:\Users\notexistentuser\.cryptag\data`,
		},
	}

	for _, tt := range tests {
		gotCollapsed, _ := Collapse(tt.path)
		assert.Equal(t, tt.wantCollapsed, gotCollapsed)
	}
}
