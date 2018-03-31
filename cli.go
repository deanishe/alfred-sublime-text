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

var (
	iconReload          = &aw.Icon{Value: "icons/reload.png"}
	iconDocs            = &aw.Icon{Value: "icons/docs.png"}
	iconHelp            = &aw.Icon{Value: "icons/help.png"}
	iconIssue           = &aw.Icon{Value: "icons/issue.png"}
	iconOff             = &aw.Icon{Value: "icons/off.png"}
	iconOn              = &aw.Icon{Value: "icons/on.png"}
	iconTrash           = &aw.Icon{Value: "icons/trash.png"}
	iconUpdateAvailable = &aw.Icon{Value: "icons/update-available.png"}
	iconUpdateOK        = &aw.Icon{Value: "icons/update-ok.png"}
	iconURL             = &aw.Icon{Value: "icons/url.png"}
)

func runConfig() {
	log.Printf(`filtering config "%s" ...`, query)
}

// Scan for projects and cache results
func runScan() {
	wf.TextErrors = true

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

	sm := NewScanManager(conf)
	if err := sm.Scan(); err != nil {
		wf.FatalError(err)
	}
}

func runSearch() {

	var projs []Project
	sm := NewScanManager(conf)

	if query != "" {
		log.Printf(`searching for "%s" ...`, query)
	}

	// Run "alfsubl rescan" in background if need be
	if sm.ScanDue() && !aw.IsRunning("rescan") {
		log.Println("rescanning for projects ...")
		cmd := exec.Command(os.Args[0], "rescan")
		if err := aw.RunInBackground("rescan", cmd); err != nil {
			log.Printf(`error running "%s rescan": %v`, os.Args[0], err)
			wf.Fatal("Error scanning for repos. See log file.")
		}
	}

	// Load data
	projs, err := sm.Load()
	if err != nil {
		wf.FatalError(err)
	}

	if len(projs) == 0 && aw.IsRunning("rescan") {

		wf.Rerun(0.3)

		wf.NewItem("Loading projects…").
			Subtitle("Results will refresh in a few seconds").
			Valid(false).
			Icon(iconReload)

		wf.SendFeedback()
		return
	}

	for _, proj := range projs {
		it := wf.NewItem(proj.Name()).
			Subtitle(util.PrettyPath(proj.Path)).
			Valid(true).
			Arg(proj.Path).
			UID(proj.Path).
			IsFile(true)

		if len(proj.Folders) > 0 {
			it.NewModifier("cmd").
				Subtitle("Open project folder").
				Arg(proj.Folders[0]).
				Valid(true)
		}
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
