//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-27
//

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/util"
)

var (
	opts = &options{}
	cli  = flag.NewFlagSet("alfsubl", flag.ContinueOnError)

	// Candidate paths to `subl` command-line program. We'll open projects
	// via `subl` because it correctly loads the workspace. Opening a
	// project with "Sublime Text.app" doesn't.
	sublPaths = []string{
		"/usr/local/bin/subl",
		"/Applications/Sublime Text 4.app/Contents/SharedSupport/bin/subl",
		"/Applications/Sublime Text.app/Contents/SharedSupport/bin/subl",
	}
	// Candidate paths to `code` command-line program.
	codePaths = []string{
		"/usr/local/bin/code",
		"/Applications/VSCodium.app/Contents/Resources/app/bin/code",
		"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
	}
)

// CLI flags
type options struct {
	// Commands
	Search      bool
	Config      bool
	Ignore      bool
	Open        bool
	OpenProject bool
	OpenFolder  bool
	Rescan      bool

	// Options
	Force bool

	// Arguments
	Query string
}

func init() {
	cli.BoolVar(&opts.Search, "search", false, "search projects")
	cli.BoolVar(&opts.Config, "conf", false, "show/filter configuration")
	cli.BoolVar(&opts.Open, "open", false, "open specified file in default app")
	cli.BoolVar(&opts.OpenProject, "project", false, "open specified project")
	cli.BoolVar(&opts.OpenFolder, "folder", false, "open specified project")
	cli.BoolVar(&opts.Rescan, "rescan", false, "re-scan for projects")
	cli.BoolVar(&opts.Force, "force", false, "force rescan")
	cli.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: alfred-sublime [options] [arguments]

Alfred workflow to show Sublime Text/VSCode projects.

Usage:
    alfred-sublime -search [<query>]
    alfred-sublime -conf [<query>]
    alfred-sublime -open <path>
    alfred-sublime -project <project file>
    alfred-sublime -folder <project file>
    alfred-sublime -rescan [-force]
    alfred-sublime -h|-help

Options:
`)

		cli.PrintDefaults()
	}
}

func commandForProject(path string) *exec.Cmd {
	var (
		app   = "Sublime Text"
		progs = sublPaths
	)
	if conf.VSCode {
		app = "Visual Studio Code"
		progs = codePaths
	}

	for _, p := range progs {
		if util.PathExists(p) {
			return exec.Command(p, path)
		}
	}

	return exec.Command("/usr/bin/open", "-a", app, path)
}

// Open a project file. CLI programs `subl` or `code` are preferred.
// If they can't be found application "Sublime Text.app" or
// "Visual Studio Code.app" is called instead.
func runOpenProject() {
	wf.Configure(aw.TextErrors(true))

	log.Printf("opening project %q ...", opts.Query)
	cmd := commandForProject(opts.Query)

	if _, err := util.RunCmd(cmd); err != nil {
		wf.Fatalf("exec command %#v: %v", cmd, err)
	}
}

// Open a project's folders
func runOpenFolder() {
	wf.Configure(aw.TextErrors(true))

	var (
		projs []Project
		sm    = NewScanManager(conf)
		err   error
	)

	if projs, err = sm.Load(); err != nil {
		wf.Fatalf("load projects: %v", err)
	}

	for _, proj := range projs {
		if proj.Path == opts.Query {
			for _, path := range proj.Folders {

				log.Printf("opening folder %q ...", path)
				cmd := exec.Command("/usr/bin/open", path)

				if _, err := util.RunCmd(cmd); err != nil {
					wf.Fatalf("run command %#v: %v", cmd, err)
				}
			}
			return
		}
	}

	wf.Fatalf("no folders found for project %q", opts.Query)
}

// Filter configuration in Alfred
func runConfig() {

	// prevent Alfred from re-ordering results
	if opts.Query == "" {
		wf.Configure(aw.SuppressUIDs(true))
	} else {
		wf.Var("query", opts.Query)
	}

	log.Printf("filtering config %q ...", opts.Query)

	if wf.UpdateAvailable() {
		wf.NewItem("Workflow Update Available").
			Subtitle("↩ or ⇥ to install update").
			Valid(false).
			UID("update").
			Autocomplete("workflow:update").
			Icon(iconUpdateAvailable)
	} else {
		wf.NewItem("Workflow Is Up To Date").
			Subtitle("↩ or ⇥ to check for update now").
			Valid(false).
			UID("update").
			Autocomplete("workflow:update").
			Icon(iconUpdateOK)
	}

	wf.NewItem("Edit Config File").
		Subtitle("Edit directories to scan").
		Valid(true).
		Arg("-open", "--", configFile).
		UID("config").
		Icon(iconSettings).
		Var("hide_alfred", "true")

	wf.NewItem("View Help File").
		Subtitle("Open workflow help in your browser").
		Arg("-open", "README.html").
		UID("help").
		Valid(true).
		Icon(iconHelp).
		Var("hide_alfred", "true")

	wf.NewItem("Report Issue").
		Subtitle("Open workflow issue tracker in your browser").
		Arg("-open", issueTrackerURL).
		UID("issue").
		Valid(true).
		Icon(iconIssue).
		Var("hide_alfred", "true")

	wf.NewItem("Visit Forum Thread").
		Subtitle("Open workflow thread on alfredforum.com in your browser").
		Arg("-open", forumThreadURL).
		UID("forum").
		Valid(true).
		Icon(iconURL).
		Var("hide_alfred", "true")

	// TODO: add "back to" setting for rescan command
	wf.NewItem("Rescan Projects").
		Subtitle("Rebuild cached list of projects").
		Arg("-rescan", "-force").
		Valid(true).
		UID("rescan").
		Icon(iconReload).
		// Var("hide_alfred", "false").
		Var("notification", "Reloading project list…").
		Var("trigger", "config")

	if opts.Query != "" {
		wf.Filter(opts.Query)
		addNavigationItems(opts.Query, "config", "rescan")
	}

	wf.WarnEmpty("No Matching Items", "Try a different query")
	wf.SendFeedback()
}

// Scan for projects and cache results
func runScan() {
	wf.Configure(aw.TextErrors(true))

	if opts.Force {
		if conf.FindInterval != 0 {
			conf.FindInterval = time.Nanosecond
		}
		if conf.MDFindInterval != 0 {
			conf.MDFindInterval = time.Nanosecond
		}
		if conf.LocateInterval != 0 {
			conf.LocateInterval = time.Nanosecond
		}
	}

	sm := NewScanManager(conf)
	if err := sm.Scan(); err != nil {
		wf.FatalError(err)
	}
	fmt.Print("Project scan completed")
}

// Open path/URL
func runOpen() {
	wf.Configure(aw.TextErrors(true))

	var args []string
	args = append(args, opts.Query)

	cmd := exec.Command("open", args...)
	if _, err := util.RunCmd(cmd); err != nil {
		wf.Fatalf("open %q: %v", opts.Query, err)
	}
}

// Filter Sublime projects in Alfred
func runSearch() {
	var (
		projs []Project
		err   error
		sm    = NewScanManager(conf)
	)

	if opts.Query != "" {
		log.Printf(`searching for "%s" ...`, opts.Query)
	}

	// Run "alfsubl rescan" in background if need be
	if sm.ScanDue() && !wf.IsRunning("rescan") {
		log.Println("rescanning for projects ...")
		cmd := exec.Command(os.Args[0], "-rescan")
		if err := wf.RunInBackground("rescan", cmd); err != nil {
			log.Printf("error running rescan: %v", err)
			wf.Fatal("Error scanning for repos. See log file.")
		}
	}

	// Load data
	if projs, err = sm.Load(); err != nil {
		wf.FatalError(err)
	}

	if len(projs) == 0 && wf.IsRunning("rescan") {

		wf.Rerun(0.1)
		wf.NewItem("Scanning projects…").
			Subtitle("Results will refresh in a few seconds").
			Valid(false).
			Icon(iconSpinner())

		wf.SendFeedback()
		return
	}

	icon := iconSublime
	if conf.VSCode {
		icon = iconVSCode
	}

	for _, proj := range projs {

		it := wf.NewItem(proj.Name()).
			Subtitle(util.PrettyPath(proj.Path)).
			Valid(true).
			Arg("-project", "--", proj.Path).
			UID(proj.Path).
			IsFile(true).
			Icon(icon).
			Var("hide_alfred", "true")

		if len(proj.Folders) > 0 {
			sub := "Open Project Folder"
			if len(proj.Folders) > 1 {
				sub += "s"
			}
			it.NewModifier("cmd").
				Subtitle(sub).
				Icon(&aw.Icon{Value: "public.folder", Type: "filetype"}).
				Arg("-folder", proj.Path)
		}
	}

	if opts.Query != "" {
		res := wf.Filter(opts.Query)
		for _, r := range res {
			log.Printf("[search] %6.2f %#v", r.Score, r.SortKey)
		}
		addNavigationItems(opts.Query, "search")
	}

	wf.WarnEmpty("No Projects Found", "Try a different query?")
	wf.SendFeedback()
}

func addNavigationItems(query string, ignore ...string) {
	if len(query) < 3 {
		return
	}
	var (
		items = []struct {
			keywords []string
			trigger  string
			title    string
			subtitle string
			arg      []string
			note     string
			icon     *aw.Icon
		}{
			{
				[]string{"reload", "rescan"},
				"rescan",
				"Rescan Projects",
				"Rescan disk & update cached list of projects",
				[]string{"-rescan", "-force"},
				"Reloading project list …",
				iconReload,
			},
			{
				[]string{"config", "prefs", "settings"},
				"config",
				"Workflow Settings",
				"Access workflow's preferences",
				nil,
				"",
				iconSettings,
			},
			{
				[]string{"search", "projects", ".st"},
				"search",
				"Search Projects",
				"Search scanned projects",
				nil,
				"",
				aw.IconWorkflow,
			},
		}
	)

	query = strings.ToLower(query)
	for _, conf := range items {
		if sliceContains(ignore, conf.trigger) {
			continue
		}
		for _, kw := range conf.keywords {
			if strings.HasPrefix(strings.ToLower(kw), query) {
				it := wf.NewItem(conf.title).
					Subtitle(conf.subtitle).
					Icon(conf.icon).
					UID("navigation-action."+conf.trigger).
					Valid(true).
					Var("trigger", conf.trigger).
					Var("query", "")

				if conf.trigger == "rescan" {
					it.Var("trigger", "search")
				}

				if conf.arg != nil {
					it.Arg(conf.arg...)
				}

				if conf.note != "" {
					it.Var("notification", conf.note)
				}
				break
			}
		}
	}
}

func sliceContains(sl []string, s string) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}
	return false
}
