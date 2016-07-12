// Steve Phillips / elimisteve
// 2012.09.26

package fun

import (
	"math/rand"
	"time"
)

func RandStrOfLen(length int, charset string) (str string) {
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)
	for i := 0; i < length; i++ {
		ndx := r.Intn(len(charset))
		str += string(charset[ndx])
	}
	return
}
