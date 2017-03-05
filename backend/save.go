// Steve Phillips / elimisteve
// 2017.03.04

package backend

import (
	"fmt"

	"github.com/cryptag/cryptag"
)

// Save turns bk into a *Config then saves it to disk.
func Save(bk Backend) error {
	cfg, err := bk.ToConfig()
	if err != nil {
		return fmt.Errorf("Error converting Backend `%s` to Config: %v",
			bk.Name(), err)
	}

	err = cfg.Save(cryptag.BackendPath)
	if err != nil {
		return fmt.Errorf("Error saving backend config to disk: %v", err)
	}

	return nil
}
