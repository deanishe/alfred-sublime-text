//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-26
//

package main

import "testing"

func TestRelDepth(t *testing.T) {
	data := []struct {
		base, dir string
		depth     int
	}{
		{"", "", 0},
		{".", ".", 0},
		{".", ".", 0},
		{"/", "/", 0},

		{"/", "/dir1", 1},
		{"/", "/dir1/dir2", 2},
		{"/", "/dir1/dir2/dir3", 3},
		{"/", "/dir1/dir2/dir3/dir4", 4},
		{"/dir1", "/dir1/dir2/dir3/dir4", 3},
		{"/dir1/dir2", "/dir1/dir2/dir3/dir4", 2},
		{"/dir1/dir2/dir3", "/dir1/dir2/dir3/dir4", 1},

		{"", "dir1", 1},
		{"", "dir1/dir2", 2},
		{"", "dir1/dir2/dir3", 3},
		{"", "dir1/dir2/dir3/dir4", 4},
		{"dir1", "dir1/dir2/dir3/dir4", 3},
		{"dir1/dir2", "dir1/dir2/dir3/dir4", 2},
		{"dir1/dir2/dir3", "dir1/dir2/dir3/dir4", 1},

		{"/dir1", "/dir2", -1},
		{"/dir1", "/", -1},
	}

	for _, td := range data {
		n := reldepth(td.base, td.dir)
		if n != td.depth {
			t.Errorf("Bad Depth. Expected=%d, Got=%d, Base=%s, Dir=%s",
				td.depth, n, td.base, td.dir)
		}
	}
}
