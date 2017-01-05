// Steven Phillips / elimisteve
// 2016.06.05

package keyutil

import (
	"fmt"
	"regexp"
	"strconv"
)

var keyRegex = regexp.MustCompile(`(\d+)`)

const validKeyLength = 32

// Parse takes the string representation of a crypto key
// (comma-separated numbers) and parses it into a usable key.
func Parse(cliDigits string) (*[32]byte, error) {
	// Pluck out all digit sequences, convert to numbers
	nums := keyRegex.FindAllString(cliDigits, -1)
	if len(nums) != validKeyLength {
		return nil, fmt.Errorf("Key must include %d numbers, not %d",
			validKeyLength, len(nums))
	}

	var newKey [32]byte

	for i := 0; i < len(newKey); i++ {
		// Parse string of numbers as (8-bit) bytes represented in base 10
		n, err := strconv.ParseUint(nums[i], 10, 8)
		if err != nil {
			return nil, fmt.Errorf("Number #%d '%v' was invalid: %v\n", i+1,
				nums[i], err)
		}
		newKey[i] = byte(n)
	}

	return &newKey, nil
}
