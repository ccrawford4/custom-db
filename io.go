package main

import (
	"fmt"
	"os"
)

func SaveData(path string, data []byte) error {
	tmp := fmt.Sprintf("%s.tmp", path)
	fp, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() {
		fp.Close()
		if err != nil {
			os.Remove(tmp)
		}
	}()

	if _, err = fp.Write(data); err != nil {
		return err
	}
	if err = fp.Sync(); err != nil {
		return err
	}
	err = os.Rename(tmp, path) // as of here is not atomic

	// MISSING:
	// 1. Should re-open the directory
	// 2. perform another sync
	return err
}
