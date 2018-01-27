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
	docopt "github.com/docopt/docopt.go"
)

var usage = `alfsubl <command> [<args>]

Alfred workflow to show Sublime Text projects.

Usage:
    alfsubl search [<query>]
    alfsubl config [<query>]
    alfsubl rescan [--force]
    alfsubl reset
    alfsubl [help]

Options:
    -f, --force    ignore cached data
    -h, --help     show this message and exit
    -v, --version  show version number and exit
`

var (
	cacheKey   = "projects.json"
	conf       *config
	configFile string
	command    string
	force      bool
	query      string
	commands   = []string{"config", "search", "rescan", "reset", "help"}
	wf         *aw.Workflow
)

func init() {
	wf = aw.New()
	configFile = filepath.Join(wf.DataDir(), "sublime.toml")
}

// set variables based on user input
func parseArgs(argv []string) error {
	opts, err := docopt.ParseArgs(usage, argv, wf.Version())
	if err != nil {
		return err
	}

	for _, s := range commands {
		if opts[s] == true {
			command = s
			break
		}
	}

	force = opts["--force"].(bool)

	if s, ok := opts["<query>"].(string); ok {
		query = s
	}

	log.Printf("opts=%#v", opts)

	return nil
}

// workflow entry point
func run() {

	var err error

	// Command-line args
	if err := parseArgs(wf.Args()); err != nil {
		log.Printf("couldn't parse args (%#v): %v", wf.Args(), err)
		wf.Fatal("Couldn't parse args. Check log file.")
	}

	// Load configuration file
	if err := initConfig(); err != nil {
		log.Printf("couldn't create config (%s): %v", configFile, err)
		wf.Fatal("Couldn't create config. Check log file.")
	}
	conf, err = loadConfig(configFile)
	if err != nil {
		log.Printf("couldn't read config (%s): %v", configFile, err)
		wf.Fatal("Couldn't read config. Check log file.")
	}

	log.Printf("command=%s, query=%s", command, query)
	// log.Printf("configFile=%s", configFile)
	// log.Printf("config=%s", conf.String())

	switch command {
	case "help", "":
		docopt.PrintHelpOnly(nil, usage)
		return

	case "search":
		runSearch()
		return

	case "config":
		runConfig()
		return

	case "rescan":
		runScan()
		return

	default:
		wf.Fatalf("Unknown Command: %s", command)
	}
}

// wrap run() in AwGo to catch and display panics
func main() {
	wf.Run(run)
}
