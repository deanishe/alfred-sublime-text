// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"log"

	aw "github.com/deanishe/awgo"
)

// Workflow icons
var (
	iconHelp            = &aw.Icon{Value: "icons/help.png"}
	iconIssue           = &aw.Icon{Value: "icons/issue.png"}
	iconLoading         = &aw.Icon{Value: "icons/loading.png"}
	iconReload          = &aw.Icon{Value: "icons/reload.png"}
	iconSettings        = &aw.Icon{Value: "icons/settings.png"}
	iconURL             = &aw.Icon{Value: "icons/url.png"}
	iconUpdateAvailable = &aw.Icon{Value: "icons/update-available.png"}
	iconUpdateOK        = &aw.Icon{Value: "icons/update-ok.png"}
	iconWarning         = &aw.Icon{Value: "icons/warning.png"}
	iconSublime         = &aw.Icon{Value: "icons/sublime.png"}
	iconVSCode          = &aw.Icon{Value: "icons/vscode.png"}
	// iconOff             = &aw.Icon{Value: "icons/off.png"}
	// iconOn              = &aw.Icon{Value: "icons/on.png"}
	// iconTrash           = &aw.Icon{Value: "icons/trash.png"}
	// iconDocs            = &aw.Icon{Value: "icons/docs.png"}
)

func init() {
	aw.IconWarning = iconWarning
}

// ReloadIcon returns a spinner icon. It rotates by 15 deg on every
// subsequent call. Use with wf.Reload(0.1) to implement an animated
// spinner.
func ReloadIcon() *aw.Icon {
	var (
		step    = 15
		max     = (45 / step) - 1
		current = wf.Config.GetInt("RELOAD_PROGRESS", 0)
		next    = current + 1
	)
	if next > max {
		next = 0
	}

	log.Printf("progress: current=%d, next=%d", current, next)

	wf.Var("RELOAD_PROGRESS", fmt.Sprintf("%d", next))

	if current == 0 {
		return iconLoading
	}

	return &aw.Icon{Value: fmt.Sprintf("icons/loading-%d.png", current*step)}
}
