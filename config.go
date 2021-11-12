//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-26
//

package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/deanishe/awgo/util"
)

const (
	// DefaultDepth is how deep to search directories by default.
	// 1 means the immediate children of the specified path, 2 means
	// its grandchildren, etc.
	DefaultDepth = 2

	// DefaultFindInterval is how often to run find
	DefaultFindInterval = 5 * time.Minute

	// DefaultMDFindInterval is how often to run mdfind
	DefaultMDFindInterval = 5 * time.Minute

	// DefaultLocateInterval is how often to run locate
	DefaultLocateInterval = 24 * time.Hour

	defaultConfig = `# How many directories deep to search by default.
# 0 = the directory itself
# 1 = immediate children of the directory
# 2 = grandchildren of the directory
# etc.
# default: 2
#
# depth = 2


# How long to cache the list of projects for.
# default: 5m
#
# cache-age = "5m"


# git-style glob patterns of paths to ignore.
# default: []
#
# E.g.:
#
# excludes = [
#   "/Applications/*",
#   "**/vim/undo/**",
# ]

# Additional paths to search with "find".
# Each search path is specified by a [[paths]] header and requires a path value.
# E.g.:
#
#  [[paths]]
#  path = "~/Dropbox"
#
# You can override the default depth:
#
#  [[paths]]
#  path = "~/Code"
#  depth = 3

`
)

var conf *config

func init() {
	conf = &config{
		Depth:          DefaultDepth,
		SearchPaths:    []*searchPath{},
		FindInterval:   DefaultFindInterval,
		MDFindInterval: DefaultMDFindInterval,
		LocateInterval: DefaultLocateInterval,
	}
}

type config struct {
	// From workflow environment variables
	FindInterval   time.Duration `toml:"-"`
	MDFindInterval time.Duration `toml:"-"`
	LocateInterval time.Duration `toml:"-"`
	VSCode         bool          `toml:"-" env:"VSCODE"`

	// From config file
	Excludes    []string      `toml:"excludes"`
	Depth       int           `toml:"depth"`
	SearchPaths []*searchPath `toml:"paths"`
}

type searchPath struct {
	Path     string   `toml:"path"`
	Excludes []string `toml:"excludes"`
	Depth    int      `toml:"depth"`
}

// Copy default settings file to data directory if there is no
// existing settings file.
func initConfig() error {
	if !util.PathExists(configFile) {
		if err := ioutil.WriteFile(configFile, []byte(defaultConfig), 0600); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}
	return nil
}

// Load configuration file.
func loadConfig(path string) (*config, error) {

	defer util.Timed(time.Now(), "load config")

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := toml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	// Load workflow variables
	if err := wf.Config.To(conf); err != nil {
		return nil, err
	}

	// Update depths and expand paths
	if conf.Depth == 0 {
		conf.Depth = DefaultDepth
	}
	for _, sp := range conf.SearchPaths {
		if sp.Depth == 0 {
			sp.Depth = conf.Depth
		}
		sp.Path = expandPath(sp.Path)
	}

	return conf, nil
}
