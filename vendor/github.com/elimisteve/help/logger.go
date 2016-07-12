// Steve Phillips / elimisteve
// 2014.11.18

package help

import (
	"errors"
	"fmt"
)

type Logger interface {
	// TODO: Add more methods as we need them
	Error(format string, args ...interface{})
}

func LogErrorf(log Logger, format string, args ...interface{}) error {
	errStr := fmt.Sprintf(format, args...)
	log.Error(errStr)
	return errors.New(errStr)
}
