//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-26
//

package main

import (
	"log"
	"path/filepath"

	aw "github.com/deanishe/awgo"
)

const (
	cacheKey        = "projects.json"
	issueTrackerURL = "https://github.com/deanishe/alfred-sublime-text/issues"
	forumThreadURL  = "https://www.alfredforum.com/topic/4510-find-and-open-sublime-text-projects/"
)

var (
	configFile string
	wf         *aw.Workflow
)

func init() {
	wf = aw.New()
	configFile = filepath.Join(wf.DataDir(), "sublime.toml")
}

// workflow entry point
func run() {

	var err error

	// Command-line args
	if err = parseArgs(wf.Args()); err != nil {
		log.Printf("couldn't parse args (%#v): %v", wf.Args(), err)
		wf.Fatal("Couldn't parse args. Check log file.")
	}

	// Load configuration file
	if err = initConfig(); err != nil {
		log.Printf("couldn't create config (%s): %v", configFile, err)
		wf.Fatal("Couldn't create config. Check log file.")
	}

	if conf, err = loadConfig(configFile); err != nil {
		log.Printf("couldn't read config (%s): %v", configFile, err)
		wf.Fatal("Couldn't read config. Check log file.")
	}

	if opts.Search {
		runSearch()
		return
	}

	if opts.Config {
		runConfig()
		return
	}

	if opts.Rescan {
		runScan()
		return
	}

	if opts.Open {
		runOpen()
		return
	}

	if opts.OpenFolder {
		runOpenFolder()
		return
	}

	if opts.OpenProject {
		runOpenProject()
		return
	}

	wf.Fatal("Unknown Command")
}

// wrap run() in AwGo to catch and display panics
func main() {
	wf.Run(run)
}
