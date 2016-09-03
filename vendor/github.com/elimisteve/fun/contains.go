// Author:  Steve Phillips / elimisteve
// Started: 2012.01.27
// Updated: 2012.05.05

package fun

import (
	"strings"
)

func ContainsAnyStrings(body string, substrings ...string) bool {
	for _, str := range substrings {
		if strings.Contains(body, str) {
			return true
		}
	}
	return false
}

func ContainsAllStrings(body string, substrings ...string) bool {
	for _, str := range substrings {
		if !strings.Contains(body, str) {
			return false
		}
	}
	return true
}

func SliceContains(slice []string, s string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}

func SliceContainsAll(slice []string, all []string) bool {
	for _, s := range all {
		if !SliceContains(slice, s) {
			return false
		}
	}
	return true
}
