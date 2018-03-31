//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-27
//

package main

import (
	"testing"
	"time"
)

var (
	testInterval = time.Second * 25
	testConf     = &config{
		FindInterval:   testInterval,
		MDFindInterval: testInterval,
		LocateInterval: testInterval,
	}
)

func TestManager(t *testing.T) {
	sm := NewScanManager(testConf)

	for _, k := range []string{"mdfind", "locate"} {

		if sm.intervals[k] != testInterval {
			t.Errorf("Bad %s interval. Expected=%v, Got=%v", k, testInterval, sm.intervals[k])
		}
	}

}
