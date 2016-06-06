// Steven Phillips / elimisteve
// 2016.06.05

package keyutil

import (
	"fmt"
	"regexp"
	"strconv"
)

var keyRegex = regexp.MustCompile(`(\d+)`)

func Parse(cliDigits string) (*[32]byte, error) {
	// Pluck out all digit sequences, convert to numbers
	nums := keyRegex.FindAllString(cliDigits, -1)
	if len(nums) != 32 {
		return nil, fmt.Errorf("Key must include 32 numbers, not %d", len(nums))
	}

	var newKey [32]byte

	for i := 0; i < len(newKey); i++ {
		n, err := strconv.ParseUint(nums[i], 10, 8)
		if err != nil {
			return nil, fmt.Errorf("Number #%d '%v' was invalid: %v\n", i+1,
				nums[i])
		}
		newKey[i] = byte(n)
	}

	return &newKey, nil
}
