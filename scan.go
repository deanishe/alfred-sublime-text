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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/deanishe/awgo/util"
	"github.com/gobwas/glob"
)

var (
	// locateDBPath = "/var/db/locate.database"
	scanners = map[string]Scanner{
		"find":   &findScanner{},
		"mdfind": &mdfindScanner{},
		"locate": &locateScanner{},
	}
)

// Scanner finds Sublime Text project files.
type Scanner interface {
	Name() string                             // name of scanner
	Scan(conf *config) (<-chan string, error) // scan for projects
}

// ScanManager loads and runs Scanners.
type ScanManager struct {
	conf      *config
	Scanners  map[string]Scanner
	intervals map[string]time.Duration
}

// NewScanManager initialises a ScanManager.
func NewScanManager(conf *config) *ScanManager {
	sm := &ScanManager{
		conf:      conf,
		Scanners:  map[string]Scanner{},
		intervals: map[string]time.Duration{},
	}

	for name, sc := range scanners {
		var d time.Duration
		switch name {
		case "mdfind":
			d = conf.MDFindInterval
		case "locate":
			d = conf.LocateInterval
		case "find":
			d = conf.FindInterval
		default:
			log.Printf("[scan] unknown scanner: %s", name)
			d = conf.FindInterval
		}
		sm.Scanners[name] = sc
		sm.intervals[name] = d
	}

	return sm
}

// ScanDue returns true if one or more scanners needs updating.
func (sm *ScanManager) ScanDue() bool {
	if !wf.Cache.Exists(cacheKey) {
		return true
	}
	if len(sm.dueScanners()) > 0 {
		log.Printf("[scan] cache expired")
		return true
	}
	return false
}

// Scan updates the cached lists of projects.
func (sm *ScanManager) Scan() error {
	var (
		due   = map[string]bool{}
		ins   []<-chan string
		out   <-chan Project
		projs []Project
		f     = &Filter{}
	)

	for _, name := range sm.dueScanners() {
		due[name] = true
	}

	for name := range sm.Scanners {
		if !sm.IsActive(name) {
			// Clear any cached results
			if err := wf.Cache.Store(sm.cacheName(name), nil); err != nil {
				log.Printf("[scan] error clearing cache: %s", err)
			}
			log.Printf("[%s] inactive", name)
			continue
		}

		if due[name] {
			sc := sm.Scanners[name]
			if c, err := sc.Scan(sm.conf); err == nil {
				log.Printf("[%s] reloading ...", name)
				ins = append(ins, cacheProjects(sm.cacheName(name), c))
			} else {
				log.Printf("[%s] error: %v", name, err)
			}
		} else {
			log.Printf("[%s] loading from cache ...", name)
			ins = append(ins, sm.scanFromCache(name))
		}
	}

	// real programs have middleware
	f.Use(makeFilterExcludes(conf.Excludes))
	f.Use(filterNotExist)
	f.Use(filterDupes)
	f.Use(filterNotProject)

	out = resultToProject(f.Apply(merge(ins...)))

	for proj := range out {
		log.Printf("[scan] project: %s (%s)", proj.Name(), util.PrettyPath(proj.Path))
		projs = append(projs, proj)
	}

	log.Printf("%d total project(s) found", len(projs))

	return wf.Cache.StoreJSON(cacheKey, projs)
}

// IsActive returns true if a scanner exists and is active.
func (sm *ScanManager) IsActive(name string) bool {
	_, ok := sm.Scanners[name]
	if !ok {
		return false
	}
	return sm.intervals[name] != 0
}

// IsDue returns true if a scanner is active and due.
func (sm *ScanManager) IsDue(name string) bool {
	if !sm.IsActive(name) {
		return false
	}

	return wf.Cache.Expired(sm.cacheName(name), sm.intervals[name])
}

// load data from cache.
func (sm *ScanManager) scanFromCache(name string) <-chan string {
	var (
		key = sm.cacheName(name)
		out = make(chan string)
	)

	go func() {
		defer close(out)
		defer util.Timed(time.Now(), fmt.Sprintf(`[cache] loaded "%s"`, name))

		if !wf.Cache.Exists(key) {
			return
		}

		data, err := wf.Cache.Load(key)
		if err != nil {
			log.Printf("[scan] error reading cache: %v", err)
			return
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			out <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Printf("[scan] error reading cache: %v", err)
		}
	}()

	return out
}

func (sm *ScanManager) dueScanners() []string {
	var (
		due   []string
		force bool
	)

	if !wf.Cache.Exists(cacheKey) {
		force = true
	}

	if age, err := wf.Cache.Age(cacheKey); err == nil {
		if fi, err := os.Stat(configFile); err == nil {
			if time.Since(fi.ModTime()) < age {
				log.Printf("[scan] config file has changed")
				force = true
			}
		}
	}

	for name := range sm.Scanners {
		if !sm.IsActive(name) {
			continue
		}

		if force || sm.IsDue(name) {
			due = append(due, name)
		}
	}
	return due
}

func (sm *ScanManager) cacheName(name string) string {
	prefix := "sublime-"
	if conf.VSCode {
		prefix = "vscode-"
	}
	return prefix + "projects-" + name + ".txt"
}

// Load loads cached Projects.
func (sm *ScanManager) Load() (projects []Project, err error) {
	if wf.Cache.Exists(cacheKey) {
		err = wf.Cache.LoadJSON(cacheKey, &projects)
	}
	return
}

// Find files with `mdfind`
type mdfindScanner struct{}

func (s *mdfindScanner) Name() string { return "mdfind" }
func (s *mdfindScanner) Scan(conf *config) (<-chan string, error) {
	cmd := exec.Command("/usr/bin/mdfind", fmt.Sprintf("kMDItemFSName == '*%s'", fileExtension))
	return lineCommand(cmd, "mdfind")
}

// Find files with `locate`
type locateScanner struct{}

func (s *locateScanner) Name() string { return "locate" }
func (s *locateScanner) Scan(conf *config) (<-chan string, error) {
	cmd := exec.Command("/usr/bin/locate", "*"+fileExtension)
	return lineCommand(cmd, "locate")
}

// Find files with `find`
type findScanner struct{}

func (s *findScanner) Name() string { return "find" }
func (s *findScanner) Scan(conf *config) (<-chan string, error) {

	var chs []<-chan string
	for _, sp := range conf.SearchPaths {
		argv := []string{sp.Path, "-maxdepth", fmt.Sprintf("%d", sp.Depth)}
		argv = append(argv, "-type", "f", "-name", "*"+fileExtension)
		ch, err := lineCommand(exec.Command("/usr/bin/find", argv...), "[find] "+sp.Path)
		if err != nil {
			return nil, err
		}
		chs = append(chs, ch)
	}

	return merge(chs...), nil
}

// Run a command and write the lines of its output to a channel.
func lineCommand(cmd *exec.Cmd, name string) (chan string, error) {

	var (
		out = make(chan string, 100)
		err error
	)

	go func() {
		defer close(out)
		defer util.Timed(time.Now(), fmt.Sprintf("%s scan", name))

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("[%s] command failed: %v", name, err)
			return
		}
		if err := cmd.Start(); err != nil {
			log.Printf("[%s] command failed: %v", name, err)
			return
		}

		// Read output and send it to channel
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			out <- scanner.Text()
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

func makeFilterExcludes(patterns []string) Filterer {
	return func(in <-chan string) <-chan string {
		return filterExcludes(in, patterns)
	}
}

// Filter files that match any of the glob patterns.
func filterExcludes(in <-chan string, patterns []string) <-chan string {
	var globs []glob.Glob

	// Compile patterns
	for _, s := range patterns {
		s = expandPath(s)
		if g, err := glob.Compile(s); err == nil {
			globs = append(globs, g)
		} else {
			log.Printf("[filter] invalid pattern (%s): %v", s, err)
		}
	}

	return filterMatches(in, func(r string) bool {
		for _, g := range globs {
			if g.Match(r) {
				log.Printf("[filter] ignored (%v): %s", g, util.PrettyPath(r))
				return true
			}
		}
		return false
	})
}

func filterNotProject(in <-chan string) <-chan string {
	return filterMatches(in, func(r string) bool {
		return !strings.HasSuffix(r, fileExtension)
	})
}

// Filter files that don't exist.
func filterNotExist(in <-chan string) <-chan string {
	return filterMatches(in, func(r string) bool {
		if _, err := os.Stat(r); err != nil {
			return true
		}
		return false
	})
}

// Filter files that have already passed through.
func filterDupes(in <-chan string) <-chan string {
	seen := map[string]bool{}
	return filterMatches(in, func(r string) bool {
		if seen[r] {
			return true
		}
		seen[r] = true
		return false
	})
}

// passes through paths from in to out, ignoring those for which ignore(path) returns true.
func filterMatches(in <-chan string, ignore func(r string) bool) <-chan string {
	var out = make(chan string)
	go func() {
		defer close(out)
		for p := range in {
			if ignore(p) {
				continue
			}
			out <- p
		}
	}()

	return out
}

func cacheProjects(key string, in <-chan string) <-chan string {

	var (
		projs = []string{}
		out   = make(chan string)
	)

	go func() {
		defer close(out)
		for p := range in {
			projs = append(projs, p)
			out <- p
		}

		sort.Strings(sort.StringSlice(projs))
		data := []byte(strings.Join(projs, "\n"))
		if err := wf.Cache.Store(key, data); err != nil {
			log.Printf("[cache] error storing %s: %v", key, err)
		} else {
			log.Printf("[cache] saved %d project(s) to %s", len(projs), key)
		}
	}()

	return out
}

// Read Sublime/VSCode project files
func resultToProject(in <-chan string) <-chan Project {
	var out = make(chan Project)

	go func() {
		defer close(out)
		for p := range in {
			proj, err := NewProject(p)
			if err != nil {
				log.Printf("[scan] couldn't read project file (%s): %v", p, err)
				continue
			}
			out <- proj
		}
	}()

	return out
}

// Combine the output of multiple channels into one.
func merge(ins ...<-chan string) <-chan string {
	var (
		wg  sync.WaitGroup
		out = make(chan string)
	)

	wg.Add(len(ins))
	for _, in := range ins {
		go func(in <-chan string) {
			defer wg.Done()
			for p := range in {
				out <- p
			}
		}(in)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
