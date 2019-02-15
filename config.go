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
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/deanishe/awgo/util"
	"github.com/pkg/errors"
)

const (
	// DefaultDepth is how deep to search directories by default.
	// 1 means the immediate children of the specified path, 2 means
	// its grandchildren, etc.
	DefaultDepth = 2

	// DefaultFindInterval is how often to run find
	DefaultFindInterval = time.Duration(5) * time.Minute

	// DefaultMDFindInterval is how often to run mdfind
	DefaultMDFindInterval = time.Duration(5) * time.Minute

	// DefaultLocateInterval is how often to run locate
	DefaultLocateInterval = time.Duration(24) * time.Hour

	envFindInterval   = "INTERVAL_FIND"
	envMDFindInterval = "INTERVAL_MDFIND"
	envLocateInterval = "INTERVAL_LOCATE"

	defaultConfig = `# How many directories deep to search by default.
# 0 = the path itself
# 1 = immediate children of the path
# 2 = grandchildren of the path
# etc.
# default: 2
# depth = 2

# How long to cache the list of projects for.
# default: 5m
# cache-age = "5m"

# Each search path is specified by a [[paths]] header and
# requires a path value.
# E.g.:
#
#  [[paths]]
#  path = "~/Dropbox"
#
# You can override the default depth:
#
#
#  [[paths]]
#  path = "~/Code"
#  depth = 3

`
)

type config struct {
	FindInterval   time.Duration `toml:"-"`
	MDFindInterval time.Duration `toml:"-"`
	LocateInterval time.Duration `toml:"-"`
	Excludes       []string      `toml:"excludes"`
	Depth          int           `toml:"depth"`
	SearchPaths    []*searchPath `toml:"paths"`
}

func (c *config) String() string {
	return fmt.Sprintf(`
INTERVAL_FIND=%s
INTERVAL_MDFIND=%s
INTERVAL_LOCATE=%s
depth=%d`, c.FindInterval, c.MDFindInterval,
		c.LocateInterval, c.Depth)
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
			return errors.Wrap(err, "write config")
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

	if conf == nil { // config file was empty
		conf = &config{
			Depth:       DefaultDepth,
			SearchPaths: []*searchPath{},
		}
	}

	// Environment variables
	conf.FindInterval = getEnvDuration(envFindInterval, 0)
	conf.MDFindInterval = getEnvDuration(envMDFindInterval, 0)
	conf.LocateInterval = getEnvDuration(envLocateInterval, 0)

	// Update depths
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

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	s := os.Getenv(key)

	log.Printf("[env] %s=%s", key, s)

	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf(`[env] invalid duration (%s) for "%s": %v`, s, key, err)
		return fallback
	}
	return d
}
