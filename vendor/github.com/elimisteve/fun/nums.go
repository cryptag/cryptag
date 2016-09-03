// Steve Phillips / elimisteve
// 2011.02.06

package fun

import (
	// "fmt"
)

// Range emulates Python's range() function. Currently only works with
// ints. TODO: Use the 'reflect' pkg to make this work for all types
func Range(values ...int) []int {
	var intSlice []int
	var min, max int
	var step int = 1

	length := len(values)
	switch length {
	default:
		// Covers length == 0
	case 1:
		max = values[0]
	case 2:
		min, max = values[0], values[1]
	case 3:
		min, max, step = values[0], values[1], values[2]
	}
	// Main loop. Gets executed no matter what
	for i := min; i < max; i += step {
		intSlice = append(intSlice, i)
	}
	return intSlice
}

// Unpacker simulates Python's list/tuple unpacking by returning the
// correct number of items -- one per element in the slice
