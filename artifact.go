package main

import (
	"os"
	"path/filepath"
)

// Artifacts retrieves files to upload.
func Artifacts(path string) ([]string, error) {
	var files []string
	file, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if file.IsDir() {
		files, err = filepath.Glob(path + "/*")
		if err != nil {
			return nil, err
		}
	} else {
		files = append(files, path)
	}

	return files, nil
}

// ArtifactNames retrieve file names to upload
func ArtifactNames(artifacts []string) []string {
	names := []string{}
	for _, a := range artifacts {
		names = append(names, filepath.Base(a))
	}
	return names
}
