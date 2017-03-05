// Steve Phillips / elimisteve
// 2017.03.04

package backend

func SandstormWebKeyToMap(webkey string) map[string]interface{} {
	return map[string]interface{}{
		"WebKey": webkey,
	}
}
