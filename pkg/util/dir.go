package util

import (
	"fmt"
	"os"
	"path/filepath"
)

func ProjectRootDirectoryMust() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("failed to get current working directory: %v", err))
	}
	wd, err = filepath.Abs(wd)
	if err != nil {
		panic(fmt.Errorf("failed to get absolute path of current working directory: %v", err))
	}
	return wd
}
