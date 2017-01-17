// Steven Phillips / elimisteve
// 2015.12.23

package fun

import "os"

func GetEnvOr(envVar, defaultVal string) string {
	val := os.Getenv(envVar)
	if val != "" {
		return val
	}
	return defaultVal
}
