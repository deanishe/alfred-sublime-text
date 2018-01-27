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
	"time"
)

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
