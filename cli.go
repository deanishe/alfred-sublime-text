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
	"path/filepath"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/util"
)

var (
	opts = &options{}
	cli  = flag.NewFlagSet("alfred-sublime", flag.ContinueOnError)

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
	OpenFolders bool
	Rescan      bool
	SetConfig   string

	// Options
	Force bool

	// Arguments
	Query string
}

func init() {
	cli.BoolVar(&opts.Search, "search", false, "search projects")
	cli.BoolVar(&opts.Config, "conf", false, "show/filter configuration")
	cli.BoolVar(&opts.Open, "open", false, "open specified file in default app")
	cli.BoolVar(&opts.OpenFolders, "folders", false, "open specified project")
	cli.BoolVar(&opts.Rescan, "rescan", false, "re-scan for projects")
	cli.BoolVar(&opts.Force, "force", false, "force rescan")
	cli.StringVar(&opts.SetConfig, "set", "", "set a configuration value")
	cli.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: alfred-sublime [options] [arguments]

Alfred workflow to show Sublime Text/VSCode projects.

Usage:
    alfred-sublime <file>...
    alfred-sublime -
    alfred-sublime -search [<query>]
    alfred-sublime -conf [<query>]
    alfred-sublime -open <path>
    alfred-sublime -folders <project file>
    alfred-sublime -rescan [-force]
    alfred-sublime -set <key> <value>
    alfred-sublime -h|-help

Options:
`)

		cli.PrintDefaults()
	}
}

func openCommand(path string) *exec.Cmd {
	// name, args := appArgs()
	// return exec.Command(name, append(args, path)...)
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

// Try to open each command-line argument in turn.
// If argument is a directory, search it for a project file.
func runOpenPaths() {
	wf.Configure(aw.TextErrors(true))

	for _, path := range cli.Args() {
		cmd := openCommand(findProject(path))
		if path == "-" {
			cmd.Stdin = os.Stdin
		}

		log.Printf("opening %q ...", path)
		if _, err := util.RunCmd(cmd); err != nil {
			log.Printf("error opening %q: %v", path, err)
		}
	}
}

func findProject(dir string) string {
	fi, err := os.Stat(dir)
	if err != nil {
		log.Printf("error inspecting file %q: %v", dir, err)
		return dir
	}
	if !fi.IsDir() {
		return dir
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("error reading directory %q: %v", dir, err)
		return dir
	}
	for _, de := range files {
		if de.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(de.Name())) == fileExtension {
			return filepath.Join(dir, de.Name())
		}
	}
	return dir
}

// Open a project's folders
func runOpenFolders() {
	wf.Configure(aw.TextErrors(true))

	var (
		sm    = NewScanManager(conf)
		projs []Project
		err   error
	)
	if projs, err = sm.Load(); err != nil {
		wf.Fatalf("load projects: %v", err)
	}

	for _, proj := range projs {
		if proj.Path != opts.Query {
			continue
		}

		for _, path := range proj.Folders {
			log.Printf("opening folder %q ...", path)
			cmd := exec.Command("/usr/bin/open", path)
			if _, err := util.RunCmd(cmd); err != nil {
				log.Printf("error opening folder %q: %v", path, err)
			}
		}
		return
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

	wf.NewItem("Rescan Projects").
		Subtitle("Rebuild cached list of projects").
		Arg("-rescan", "-force").
		Valid(true).
		UID("rescan").
		Icon(iconReload).
		// Var("hide_alfred", "false").
		Var("notification", "Reloading project list…").
		Var("trigger", "config")

	wf.NewItem("Edit Config File").
		Subtitle("Edit directories to scan").
		Valid(true).
		Arg("-open", "--", configFile).
		UID("config").
		Icon(iconSettings).
		Var("hide_alfred", "true")

	v := "true"
	editor := "Sublime Text"
	other := "VS Code"
	icon := iconSublime
	if conf.VSCode {
		v = "false"
		icon = iconVSCode
		editor, other = other, editor
	}
	wf.NewItem("Editor: "+editor).
		Subtitle("↩ to switch to "+other).
		Valid(true).
		Arg("-set", "VSCODE", v).
		Icon(icon).
		Var("notification", "Using "+other)

	v = "true"
	icon = iconOff
	if conf.ActionProjectFile {
		v = "false"
		icon = iconOn
	}
	wf.NewItem("Action Project File").
		Subtitle("Action path of project file instead of first project directory").
		Valid(true).
		Arg("-set", "ACTION_PROJECT_FILE", v).
		Icon(icon)

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
		Icon(iconForum).
		Var("hide_alfred", "true")

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

// Save a config value and re-open settings view.
func runSetConfig() {
	wf.Configure(aw.TextErrors(true))

	var (
		key   = opts.SetConfig
		value = opts.Query
	)
	if err := wf.Config.Set(key, value, false).Do(); err != nil {
		wf.Fatalf("set config %q to %q: %v", key, value, err)
	}
	log.Printf("set %q to %q", key, value)
	if err := wf.Alfred.RunTrigger("config", ""); err != nil {
		wf.Fatalf("run trigger config: %v", err)
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

	// Run "alfred-sublime -rescan" in background if need be
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
			Subtitle("Results will be available shortly").
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
		path := proj.Folder()
		if conf.ActionProjectFile {
			path = proj.Path
		}
		it := wf.NewItem(proj.Name()).
			Subtitle(util.PrettyPath(path)).
			Valid(true).
			// Arg("-project", "--", proj.Path).
			Arg(proj.Path).
			IsFile(true).
			UID(proj.Path).
			Copytext(path).
			Action(path).
			Icon(icon).
			Var("hide_alfred", "true")

		if len(proj.Folders) > 0 {
			sub := "Open Project Folder"
			if len(proj.Folders) > 1 {
				sub += "s"
			}
			it.NewModifier("cmd").
				Subtitle(sub).
				Icon(&aw.Icon{Value: proj.Folder(), Type: "fileicon"}).
				Arg("-folders", proj.Path)
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

func addNavigationItems(query, backTo string, ignore ...string) {
	if len(query) < 3 {
		return
	}
	ignore = append(ignore, backTo)
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
				// Trigger doesn't exist, but we can't put the
				// real trigger (backTo) here yet because it's
				// the current action, which we want to filter out
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
			if !strings.HasPrefix(strings.ToLower(kw), query) {
				continue
			}
			it := wf.NewItem(conf.title).
				Subtitle(conf.subtitle).
				Icon(conf.icon).
				UID("navigation-action."+conf.trigger).
				Valid(true).
				Var("trigger", conf.trigger).
				Var("query", "")

			// override non-existent "rescan" trigger
			if conf.trigger == "rescan" {
				it.Var("trigger", backTo)
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

func sliceContains(sl []string, s string) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}
	return false
}
