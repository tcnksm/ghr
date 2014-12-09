package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	ID_NOT_FOUND = -1
)

func deleteRelease(info *Info) (err error) {

	if info.ID == ID_NOT_FOUND {
		return nil
	}

	fmt.Fprintf(os.Stderr, "Delete Release %d associated with Tag %s \n", info.ID, info.TagName)

	err = DeleteRelease(info)
	if err != nil {
		return err
	}

	err = DeleteTag(info)
	if err != nil {
		return err
	}

	// executing delete tag has time lag
	// So we need to wait for a while
	// This is stupid implementation...
	time.Sleep(3 * time.Second)
	info.ID = ID_NOT_FOUND

	return nil
}

func SetRelease(info *Info, replace bool) (err error) {

	info.ID, err = GetReleaseID(info)
	if err != nil {
		return err
	}

	if replace {
		if err = deleteRelease(info); err != nil {
			return err
		}
	}

	if info.ID != ID_NOT_FOUND {
		return nil
	}

	info.ID, err = CreateNewRelease(info)
	if err != nil {
		return err
	}

	if info.ID == ID_NOT_FOUND {
		return fmt.Errorf("Counld not retrieve release ID\n")
	}

	return nil
}

// Extract delete target ID
func DeleteTargetID(targets []DeleteTarget, path string) int {
	for _, t := range targets {
		if t.Name == filepath.Base(path) {
			return t.AssetId
		}
	}

	return ID_NOT_FOUND
}
