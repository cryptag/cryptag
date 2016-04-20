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
	return TimeStr(Now())
}

func TimeStr(t time.Time) string {
	y, m, d := t.Date()
	hr, min, sec := t.Clock()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d", y, m, d, hr, min, sec)
}
