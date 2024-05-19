package main

import (
	_ "embed"
	"path/filepath"
	"strings"
)

func checkExtensions(fileName string) bool {
	if len(*options.extensions) == 0 {
		return true
	}
	if options.extensions == nil {
		return true
	}
	for _, suf := range *options.extensions {
		if strings.HasSuffix(fileName, suf) {
			return true
		}
	}
	return false
}

func checkLanguages(fileName string) bool {
	if len(*options.languages) == 0 {
		return true
	}
	// //go:embed language_extensions.json
	// var file []byte
	// TODO
	return true
}

func checkExlude(fileName string) bool {
	if len(*options.exclude) == 0 {
		return true
	}
	if options.exclude == nil {
		return true
	}
	for _, pattern := range *options.exclude {
		if ok, _ := filepath.Match(pattern, fileName); ok {
			return false
		}
	}
	return true
}

func checkRestrictTo(fileName string) bool {
	if len(*options.restrictTo) == 0 {
		return true
	}
	if options.restrictTo == nil {
		return true
	}
	for _, pattern := range *options.restrictTo {
		if ok, _ := filepath.Match(pattern, fileName); ok {
			return true
		}
	}
	return false
}

func isFileSuitable(fileName string) bool {
	return checkExtensions(fileName) && checkLanguages(fileName) &&
		checkExlude(fileName) && checkRestrictTo(fileName)
}
