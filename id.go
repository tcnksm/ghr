package main

import (
	"fmt"
	"os"
	"time"
)

const (
	ID_NOT_SET = -1
)

func deleteRelease(info *Info) (err error) {

	if info.ID == ID_NOT_SET {
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
	info.ID = ID_NOT_SET

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

	if info.ID != ID_NOT_SET {
		return nil
	}

	info.ID, err = CreateNewRelease(info)
	if err != nil {
		return err
	}

	if info.ID == ID_NOT_SET {
		return fmt.Errorf("Counld not retrieve release ID\n")
	}

	return nil
}
