// Steven Phillips / elimisteve
// 2016.04.18

package cryptag

import "os"

// UseTor is one centralized place for all code to check to see if it
// should do HTTP calls over Tor.
var UseTor = false

func init() {
	if os.Getenv("TOR") == "1" {
		UseTor = true
	}
}

type CanUseTor interface {
	UseTor() error
}
