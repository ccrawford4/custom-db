package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func SaveData(path string, data []byte) error {
	dir := filepath.Dir(path)
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

	if _, err = fp.Write(data); err != nil { // save to the temporary file
		return err
	}

	if err = fp.Sync(); err != nil { // fsync
		return err
	}

	// as of here is not atomic
	if err = os.Rename(tmp, path); err != nil { // replace the target
		return err
	}

	// To make it power-atomic, we should open the directory, and do an fsync on it
	fp, err = os.Open(dir)
	if err != nil {
		return err
	}

	// fsync
	if err = fp.Sync(); err != nil {
		return err
	}

	return nil
}
