//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-26
//

package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
)

const (
	issueTrackerURL = "https://github.com/deanishe/alfred-sublime-text/issues"
	forumThreadURL  = "https://www.alfredforum.com/topic/4510-find-and-open-sublime-text-projects/"
	repo            = "deanishe/alfred-sublime-text"
)

var (
	cacheKey      = "sublime-projects.json"
	fileExtension = ".sublime-project"

	configFile string
	wf         *aw.Workflow
)

func init() {
	wf = aw.New(update.GitHub(repo), aw.HelpURL(issueTrackerURL))
	configFile = filepath.Join(wf.DataDir(), "sublime.toml")
}

// workflow entry point
func run() {
	var err error
	if err = cli.Parse(wf.Args()); err != nil {
		if err == flag.ErrHelp {
			return
		}
		wf.FatalError(err)
	}
	opts.Query = cli.Arg(0)

	// Load configuration file
	if err = initConfig(); err != nil {
		log.Printf("couldn't create config (%s): %v", configFile, err)
		wf.Fatal("Couldn't create config. Check log file.")
	}

	if conf, err = loadConfig(configFile); err != nil {
		log.Printf("couldn't read config (%s): %v", configFile, err)
		wf.Fatal("Couldn't read config. Check log file.")
	}

	log.Printf("%#v", opts)
	if wf.Debug() {
		log.Printf("args=%#v => %#v", wf.Args(), cli.Args())
		log.Print(spew.Sdump(conf))
	}

	// Naughtily switch globals to propagate VSCode mode
	if conf.VSCode {
		cacheKey = "vscode-projects.json"
		fileExtension = ".code-workspace"
	}

	if opts.Config {
		runConfig()
	} else if opts.Rescan {
		runScan()
	} else if opts.Open {
		runOpen()
	} else if opts.OpenFolder {
		runOpenFolder()
	} else if opts.OpenProject {
		runOpenProject()
	} else if opts.Search {
		runSearch()
	}
}

// wrap run() in AwGo to catch and display panics
func main() {
	wf.Run(run)
}
