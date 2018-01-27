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

var (
	totals   = map[string]int{}
	accepted = map[string][]string{}
)

func runScan() {
	wf.TextErrors = true

	var res []ScanResult

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

	for r := range scan() {
		accepted[r.Scanner] = append(accepted[r.Scanner], r.Path)
		res = append(res, r)
		// log.Println(p)
	}

	log.Printf("%d project(s)", len(res))
	for n := range totals {
		log.Printf(`[filter] %d/%d accepted in "%s"`, len(accepted[n]), totals[n], n)
		if n == "locate" {
			for _, p := range accepted[n] {
				log.Printf("  %s", p)
			}
		}
	}

	if err := wf.Cache.StoreJSON("projects.json", res); err != nil {
		wf.FatalError(err)
	}
}
