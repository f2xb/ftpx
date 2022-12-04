package xftp

import (
	"fmt"
	"path/filepath"
	"strings"
)

func checkPath(path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("dir is empty")
	}

	path = strings.Trim(path, `\`)
	path = strings.Trim(path, `/`)
	path = strings.ReplaceAll(path, `\\`, `/`)
	path = strings.ReplaceAll(path, `\`, `/`)

	dst, name := filepath.Split(path)
	if dst == "" && name != "" {
		return []string{}, nil
	}

	var dirs []string
	if strings.Contains(dst, `/`) {
		dirs = strings.Split(dst, `/`)
	}
	if len(dirs) == 0 {
		dirs = append(dirs, dst)
	}
	return dirs, nil
}
