//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-26
//

package main

import (
	"os"
	"path/filepath"
	"strings"
)

// calculate the relative depth between base and dir.
//
// base itself has a depth of 0, its immediate children of 1 etc.
// If dir is not under base (and is not base itself), -1 is returned.
func reldepth(base, dir string) int {

	base = filepath.Clean(base)
	dir = filepath.Clean(dir)

	if base == "." {
		base = ""
	}
	if dir == "." {
		dir = ""
	}

	if !strings.HasPrefix(dir, base) {
		// log("no match: base=%s, dir=%s", base, dir)
		return -1
	}

	if base == dir {
		return 0
	}

	if strings.HasPrefix(dir, "/") {
		base = base[1:]
		dir = dir[1:]
	}

	db := len(strings.Split(base, "/"))
	dd := len(strings.Split(dir, "/"))

	if base == "" {
		db = 0
	}
	if dir == "" {
		dd = 0
	}

	return (dd - db)
}

// Replace ~ in a path with the home directory.
func expandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	path = os.ExpandEnv("$HOME") + path[1:]
	return filepath.Clean(path)
}
