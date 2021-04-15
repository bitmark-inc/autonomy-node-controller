package main

import (
	"os"
)

// IsPodBound returns true if a bound file is existed
func (c *Controller) IsPodBound() (bool, error) {
	_, err := os.Stat(c.bindingFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// BindAccount writes the bound file to ensure a pod is bound
func (c *Controller) BindAccount() error {
	return os.WriteFile(c.bindingFile, []byte(""), 0600)
}
