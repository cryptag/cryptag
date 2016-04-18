// Steven Phillips / elimisteve
// 2016.04.06

package cryptag

import (
	"fmt"
	"time"
)

func Now() time.Time {
	return time.Now().UTC()
}

func NowStr() string {
	now := Now()
	y, m, d := now.Date()
	hr, min, sec := now.Clock()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d", y, m, d, hr, min, sec)
}
