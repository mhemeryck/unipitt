package unipitt

import (
	"os"
	"path/filepath"
	"regexp"
)

// findPathsByRegex find matching paths where a regex matches (on the name of a given oflder, not full path)
func findPathsByRegex(root string, pattern string) (paths []string, err error) {
	regex, err := regexp.Compile(pattern)
	// Walk the folder structure
	err = filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if regex.MatchString(info.Name()) {
				paths = append(paths, path)
			}
			return err
		})
	return
}
