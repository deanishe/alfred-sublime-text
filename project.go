//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-27
//

package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Project is a Sublime Text project.
type Project struct {
	Path    string
	Folders []string
}

type sublimeProject struct {
	Folders []sublimeFolder `json:"folders"`
}

type sublimeFolder struct {
	Path string `json:"path"`
}

// NewProject reads a .sublime-project file.
func NewProject(path string) (Project, error) {

	var (
		dir  = filepath.Dir(path)
		proj = Project{Path: path}
		raw  = sublimeProject{}
		data []byte
		err  error
	)

	data, err = ioutil.ReadFile(path)
	if err != nil {
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

	p := filepath.Join(base, relpath)
	return filepath.Clean(p)
}
