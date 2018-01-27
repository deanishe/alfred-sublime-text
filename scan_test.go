//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-27
//

package main

import "testing"

func TestScanResult(t *testing.T) {
	paths := []struct {
		in, out string
	}{
		{"", ""},
		{".", "."},
		{"path/.", "."},
		{"/", "/"},
		{"~/Documents", "Documents"},
		{"/Applications/Safari.app", "Safari"},
		{"./Alfred Sublime.sublime-project", "Alfred Sublime"},
		{"./path/to/something.txt", "something"},
	}

	for _, td := range paths {

		r := ScanResult{Path: td.in}

		if r.Name() != td.out {
			t.Errorf("Bad Name. Expected=%v, Got=%v", td.out, r.Name())
		}
	}
}
