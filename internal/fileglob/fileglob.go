package fileglob

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// Glob returns all files matching pattern relative to the current directory.
// In addition to standard filepath.Match wildcards (*, ?, [...]) it supports
// ** to match zero or more directories.
func Glob(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		return filepath.Glob(pattern)
	}

	var matches []string
	root := staticPrefix(pattern)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		matched, matchErr := Match(pattern, filepath.ToSlash(path))
		if matchErr != nil {
			return matchErr
		}
		if matched {
			matches = append(matches, path)
		}
		return nil
	})

	return matches, err
}

// Match reports whether name matches the glob pattern.
// The pattern may use ** to match zero or more path segments.
func Match(pattern, name string) (bool, error) {
	return matchParts(
		strings.Split(pattern, "/"),
		strings.Split(name, "/"),
	)
}

func matchParts(pat, name []string) (bool, error) {
	for len(pat) > 0 {
		if pat[0] == "**" {
			pat = pat[1:]

			if len(pat) == 0 {
				return true, nil
			}

			// Try the remaining pattern against every suffix of name.
			for i := 0; i <= len(name); i++ {
				if ok, err := matchParts(pat, name[i:]); err != nil {
					return false, err
				} else if ok {
					return true, nil
				}
			}
			return false, nil
		}

		if len(name) == 0 {
			return false, nil
		}

		matched, err := filepath.Match(pat[0], name[0])
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}

		pat = pat[1:]
		name = name[1:]
	}

	return len(name) == 0, nil
}

// staticPrefix returns the longest path prefix that contains no wildcards.
func staticPrefix(pattern string) string {
	parts := strings.Split(pattern, "/")
	var prefix []string
	for _, p := range parts {
		if p == "**" || strings.ContainsAny(p, "*?[") {
			break
		}
		prefix = append(prefix, p)
	}
	if len(prefix) == 0 {
		return "."
	}
	return strings.Join(prefix, "/")
}
