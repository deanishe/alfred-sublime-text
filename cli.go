//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-27
//

package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/util"
)

func runConfig() {
	log.Printf(`filtering config "%s" ...`, query)
}

// Scan for projects and cache results
func runScan() {
	wf.TextErrors = true

	var projs []Project

	if force {
		if conf.FindInterval != 0 {
			conf.FindInterval = time.Nanosecond * 1
		}
		if conf.MDFindInterval != 0 {
			conf.MDFindInterval = time.Nanosecond * 1
		}
		if conf.LocateInterval != 0 {
			conf.LocateInterval = time.Nanosecond * 1
		}
	}

	for proj := range scan() {
		projs = append(projs, proj)
		// log.Println(p)
	}

	log.Printf("%d project(s)", len(projs))

	if err := wf.Cache.StoreJSON("projects.json", projs); err != nil {
		wf.FatalError(err)
	}
}

func runSearch() {

	var projs []Project

	log.Printf(`searching for "%s" ...`, query)

	// Run "alfsubl rescan" in background if need be
	if scanDue() && !aw.IsRunning("rescan") {
		log.Println("recanning for projects ...")
		cmd := exec.Command(os.Args[0], "rescan")
		if err := aw.RunInBackground("rescan", cmd); err != nil {
			log.Printf(`error running "%s rescan": %v`, os.Args[0], err)
			wf.Fatal("Error scanning for repos. See log file.")
		}
	}

	// Load data
	if wf.Cache.Exists(cacheKey) {
		if err := wf.Cache.LoadJSON(cacheKey, &projs); err != nil {
			wf.FatalError(err)
		}
	}

	for _, proj := range projs {
		wf.NewItem(proj.Name()).
			Subtitle(util.PrettyPath(proj.Path)).
			Valid(true).
			Arg(proj.Path).
			IsFile(true)
	}

	if query != "" {
		res := wf.Filter(query)
		for _, r := range res {
			log.Printf("[search] %0.2f %#v", r.Score, r.SortKey)
		}
	}

	wf.WarnEmpty("No Projects Found", "Try a different query?")
	wf.SendFeedback()
}
