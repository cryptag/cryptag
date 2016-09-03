// Steve Phillips / elimisteve
// 2014.08.09

package help

import (
	"errors"
	"fmt"
)

func TxError(rollbackErr error, format string, params ...interface{}) error {
	errStr := fmt.Sprintf(format, params...)
	if rollbackErr != nil {
		errStr += "... and the rollback failed with: " + rollbackErr.Error()
	}
	return errors.New(errStr)
}
