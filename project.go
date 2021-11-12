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
	"path/filepath"
	"strings"

	// Supports comments in JSON, which is required to read
	// Sublime Text or VS Code project files.
	json "github.com/yosuke-furukawa/json5/encoding/json5"
)

// Project is a Sublime Text or VS Code project.
type Project struct {
	Path    string // to project file
	Folders []string
}

// Folder returns the path of the first project folder, falling
// back to the path of the folder containing the project file.
func (p Project) Folder() string {
	if len(p.Folders) == 0 {
		return filepath.Dir(p.Path)
	}
	return p.Folders[0]
}

// Name returns the name of the project (the filename w/o extension).
func (p Project) Name() string {

	if p.Path == "" {
		return ""
	}

	s, x := filepath.Base(p.Path), filepath.Ext(p.Path)
	if x == "" || x == "." {
		return s
	}

	return s[0 : len(s)-len(x)]
}

type sublimeProject struct {
	Folders []sublimeFolder `json:"folders"`
}

type sublimeFolder struct {
	Path string `json:"path"`
}

// NewProject reads a .sublime-project or .code-workspace file.
func NewProject(path string) (Project, error) {
	var (
		dir  = filepath.Dir(path)
		proj = Project{Path: path}
		raw  = sublimeProject{}
		data []byte
		err  error
	)

	if data, err = ioutil.ReadFile(path); err != nil {
		return proj, err
	}

	if err = json.Unmarshal(data, &raw); err == nil {
		proj.Folders = []string{}
		for _, f := range raw.Folders {
			if p := resolvePath(dir, f.Path); p != "" {
				proj.Folders = append(proj.Folders, p)
			}
		}
	}
	return proj, err
}

func resolvePath(base, relpath string) string {
	if strings.HasPrefix(relpath, "/") {
		return relpath
	}
	if base == "" || relpath == "" {
		return ""
	}

	return filepath.Clean(filepath.Join(base, relpath))
}
