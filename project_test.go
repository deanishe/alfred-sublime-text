//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-27
//

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	testProjJS = `{
	"folders":
	[
		{
			"path": "/usr/local/bin"
		},
		{
			"path": "/etc"
		},
		{
			"path": "."
		}
	]
}`
	testProjPaths = []string{"/usr/local/bin", "/etc"}
)

func withTestFile(data []byte, fn func(path string)) error {
	f, err := ioutil.TempFile("", "alfred-sublime-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(data); err != nil {
		return err
	}

	fn(f.Name())

	return nil
}

func TestParseProject(t *testing.T) {
	err := withTestFile([]byte(testProjJS), func(path string) {

		dir := filepath.Dir(path)
		paths := make([]string, len(testProjPaths))
		copy(paths, testProjPaths)
		paths = append(paths, dir)

		proj, err := NewProject(path)
		if err != nil {
			t.Fatalf("couldn't create new project: %v", err)
		}

		if proj.Path != path {
			t.Errorf("Bad Path. Expected=%v, Got=%v", path, proj.Path)
		}

		if len(proj.Folders) != len(paths) {
			t.Fatalf("Bad Folders length. Expected=%v, Got=%v", len(paths), len(proj.Folders))
		}

		for i, s := range proj.Folders {
			if s != paths[i] {
				t.Errorf("Bad Folder. Expected=%v, Got=%v", paths[i], s)
			}
		}

		if s := proj.Folder(); s != paths[0] {
			t.Errorf("Bad Folder. Expected=%v, Got=%v", paths[0], s)
		}

	})
	if err != nil {
		t.Fatalf("couldn't create tempfile: %v", err)
	}

}

func TestResolvePath(t *testing.T) {
	data := []struct {
		base, rel, out string
	}{
		{"/", "home/bob", "/home/bob"},
		{"/home/bob", ".", "/home/bob"},
		{".", "/home/bob", "/home/bob"},
		{".", "bob", "bob"},
		{".", "bob/public", "bob/public"},
		{"./bob", "public", "bob/public"},
		{"home", "bob", "home/bob"},
		{"", "", ""},
		{"home", "", ""},
		{"", "bob", ""},
	}

	for _, td := range data {
		s := resolvePath(td.base, td.rel)
		if s != td.out {
			t.Errorf("Bad ResolvePath. Expected=%v, Got=%v", td.out, s)
		}
	}
}

func TestProjectNames(t *testing.T) {

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
		proj := Project{Path: td.in}
		if proj.Name() != td.out {
			t.Errorf("Bad Name. Expected=%v, Got=%v", td.out, proj.Name())
		}
	}
}
