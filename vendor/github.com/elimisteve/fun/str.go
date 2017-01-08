// Steve Phillips / elimisteve
// 2014.01.05

package fun

import (
	"github.com/dchest/uniuri"
)

// RandomString produces a string of length `length` using characters
// from `alphabet`
func RandomString(alphabet string, length int) string {
	return uniuri.NewLenChars(length, []byte(alphabet))
}
