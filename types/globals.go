// Steve Phillips / elimisteve
// 2015.02.24

package types

import "os"

var (
	Debug = false
)

func init() {
	if os.Getenv("DEBUG") == "1" {
		Debug = true
	}
}
