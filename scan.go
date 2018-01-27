//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-26
//

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/deanishe/awgo/util"
	"github.com/gobwas/glob"
)

var (
	locateDBPath = "/var/db/locate.database"
	scanners     = map[string]Scanner{
		"mdfind": &mdfindScanner{&cacher{name: "mdfind"}},
		"locate": &locateScanner{&cacher{name: "locate"}},
	}
)

// return true if at least one scanner wants to scan
func scanDue() bool {

	if !wf.Cache.Exists(cacheKey) {
		return true
	}

	for _, sc := range scanners {
		if sc.Due() {
			log.Printf("[%s] rescan due", sc.Name())
			return true
		}
	}
	return false
}

// load all ST projects
func scan() <-chan ScanResult {

	var (
		ins []<-chan ScanResult
		out = make(<-chan ScanResult)
	)

	for name, scanner := range scanners {

		if !scanner.Ready() {
			log.Printf("[%s] inactive", scanner.Name())
			continue
		}

		log.Printf("[%s] starting ...", name)
		if c, err := scanner.Scan(); err == nil {
			ins = append(ins, c)
		} else {
			log.Printf("[%s] error: %v", name, err)
		}
	}

	// real programs have middleware
	out = filterExcludes(
		filterNotExist(
			filterDupes(
				filterNotProject(
					merge(ins...),
				),
			),
		),
		conf.Excludes)

	return out
}

// ScanResult is returned by a scanner.
type ScanResult struct {
	Dir     string // first directory listed in the .sublime-project file
	Path    string // path of the .sublime-project file
	Scanner string // name of scanner that found it
}

// Name returns the name of the project (the filename w/o extension).
func (r ScanResult) Name() string {

	if r.Path == "" {
		return ""
	}

	s, x := filepath.Base(r.Path), filepath.Ext(r.Path)
	if x == "" || x == "." {
		return s
	}

	return s[0 : len(s)-len(x)]
}

func (r ScanResult) String() string {
	return fmt.Sprintf("[%s] %s", r.Scanner, r.Path)
}

// Scanner finds Sublime Text project files.
type Scanner interface {
	Name() string                     // name of scanner
	Due() bool                        // whether scanner wants to rescan
	Ready() bool                      // whether scanner is runnable
	Scan() (<-chan ScanResult, error) // scan for projects
}

// cacher is a base Scanner that can load and save cached data.
type cacher struct {
	name      string
	fromCache bool
}

func (c *cacher) Name() string { return c.name }

func (c *cacher) cacheName() string {
	return "projects-" + c.Name() + ".txt"
}

// HasCache returns true if cache is valid.
func (c *cacher) HasCache(maxAge time.Duration) bool {
	return !wf.Cache.Expired(c.cacheName(), maxAge)
}

func (c *cacher) Loader() chan ScanResult {

	var out = make(chan ScanResult)

	go func() {

		defer close(out)

		data, err := wf.Cache.Load(c.cacheName())
		if err != nil {
			log.Printf(`[cache] load error for "%s": %v`, c.Name(), err)
			return
		}

		buf := bytes.NewBuffer(data)
		scanner := bufio.NewScanner(buf)
		var i int
		for scanner.Scan() {
			out <- ScanResult{Path: scanner.Text(), Scanner: c.Name()}
			i++
		}
		if err := scanner.Err(); err != nil {
			log.Printf(`[cache] reading error for "%s": %v`, c.Name(), err)
		} else {
			log.Printf(`[cache] %d projects loaded for "%s"`, i, c.Name())
		}
	}()

	return out
}
func (c *cacher) Saver(in <-chan ScanResult, err error) (chan ScanResult, error) {

	if err != nil {
		return nil, err
	}

	var (
		out   = make(chan ScanResult)
		paths []string
	)

	go func() {
		defer close(out)

		for r := range in {

			out <- r

			paths = append(paths, r.Path)
		}

		data := []byte(strings.Join(paths, "\n") + "\n")
		if err := wf.Cache.Store(c.cacheName(), data); err != nil {
			log.Printf(`[cache] save error for "%s": %v`, c.Name(), err)
			return
		}
		log.Printf(`[cache] %d projects saved for "%s"`, len(paths), c.Name())
	}()

	return out, nil
}

// Find .sublime-project files with `mdfind`
type mdfindScanner struct {
	*cacher
}

func (s *mdfindScanner) Name() string { return "mdfind" }

func (s *mdfindScanner) Due() bool {
	if conf.MDFindInterval == 0 {
		return false
	}
	return !s.HasCache(conf.MDFindInterval)
}

func (s *mdfindScanner) Ready() bool {
	return conf.MDFindInterval != 0
}

func (s *mdfindScanner) Scan() (<-chan ScanResult, error) {
	if s.HasCache(conf.MDFindInterval) {
		return s.Loader(), nil
	}
	cmd := exec.Command("/usr/bin/mdfind", "-name", ".sublime-project")
	return s.Saver(lineCommand(cmd, s.Name()))
}

// Find *.sublime-project files with `locate`
type locateScanner struct {
	*cacher
}

func (s *locateScanner) Name() string { return "locate" }

func (s *locateScanner) Due() bool {
	if conf.LocateInterval == 0 {
		return false
	}
	return !s.HasCache(conf.LocateInterval)
}

func (s *locateScanner) Ready() bool {
	if conf.LocateInterval == 0 {
		return false
	}
	if !util.PathExists(locateDBPath) {
		return false
	}
	return true
}
func (s *locateScanner) Scan() (<-chan ScanResult, error) {
	if s.HasCache(conf.LocateInterval) {
		return s.Loader(), nil
	}
	cmd := exec.Command("/usr/bin/locate", "*.sublime-project")
	return s.Saver(lineCommand(cmd, s.Name()))
}

// Run a command and write the lines of its output to a channel.
func lineCommand(cmd *exec.Cmd, name string) (chan ScanResult, error) {

	var (
		out = make(chan ScanResult, 100)
		err error
	)

	go func() {

		defer close(out)
		defer timed(time.Now(), fmt.Sprintf("%s scan", name))

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("[%s] command failed: %v", name, err)
			return
		}

		if err := cmd.Start(); err != nil {
			log.Printf("[%s] command failed: %v", name, err)
			return
		}

		// Read mdfind output and send it to channel
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			out <- ScanResult{Path: scanner.Text(), Scanner: name}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("[%s] couldn't parse output: %v", name, err)
		}

		if err != cmd.Wait() {
			log.Printf("[%s] command failed: %v", name, err)
		}
	}()

	return out, err
}

// Filter files that match any of the glob patterns.
func filterExcludes(in <-chan ScanResult, patterns []string) <-chan ScanResult {
	var globs []glob.Glob

	// Compile patterns
	for _, s := range patterns {
		if g, err := glob.Compile(s); err == nil {
			globs = append(globs, g)
		} else {
			log.Printf("[filter] invalid pattern (%s): %v", s, err)
		}
	}

	return filterMatches(in, func(r ScanResult) bool {
		for _, g := range globs {
			if g.Match(r.Path) {
				// log.Printf("[filter] ignored (%s): %s", g, r.String())
				return true
			}
		}
		return false
	})
}

func filterNotProject(in <-chan ScanResult) <-chan ScanResult {
	return filterMatches(in, func(r ScanResult) bool {
		return !strings.HasSuffix(r.Path, ".sublime-project")
	})
}

// Filter files that don't exist.
func filterNotExist(in <-chan ScanResult) <-chan ScanResult {
	return filterMatches(in, func(r ScanResult) bool {
		if _, err := os.Stat(r.Path); err != nil {
			// log.Printf("[filter] doesn't exist: %s", p)
			return true
		}
		return false
	})
}

// Filter files that have already passed through.
func filterDupes(in <-chan ScanResult) <-chan ScanResult {

	seen := map[string]bool{}

	return filterMatches(in, func(r ScanResult) bool {

		if seen[r.Path] {
			// log.Printf("[filter] duplicate: %s", r.String())
			return true
		}

		seen[r.Path] = true
		return false
	})
}

// passes through paths from in to out, ignoring those for which ignore(path) returns true.
func filterMatches(in <-chan ScanResult, ignore func(r ScanResult) bool) <-chan ScanResult {

	var out = make(chan ScanResult)

	go func() {
		defer close(out)

		for r := range in {
			if ignore(r) {
				continue
			}
			out <- r
		}
	}()

	return out
}

// Combine the output of multiple channels into one.
func merge(ins ...<-chan ScanResult) <-chan ScanResult {
	var (
		wg  sync.WaitGroup
		out = make(chan ScanResult)
	)

	wg.Add(len(ins))

	for _, in := range ins {

		go func(in <-chan ScanResult) {
			defer wg.Done()
			for r := range in {
				out <- r
			}
		}(in)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
