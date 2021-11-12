// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package main

import (
	"fmt"

	aw "github.com/deanishe/awgo"
)

// Workflow icons
var (
	iconError           = &aw.Icon{Value: "icons/error.png"}
	iconForum           = &aw.Icon{Value: "icons/forum.png"}
	iconHelp            = &aw.Icon{Value: "icons/help.png"}
	iconIssue           = &aw.Icon{Value: "icons/issue.png"}
	iconReload          = &aw.Icon{Value: "icons/reload.png"}
	iconOn              = &aw.Icon{Value: "icons/toggle-on.png"}
	iconOff             = &aw.Icon{Value: "icons/toggle-off.png"}
	iconSettings        = &aw.Icon{Value: "icons/settings.png"}
	iconSublime         = &aw.Icon{Value: "icons/sublime.png"}
	iconUpdateAvailable = &aw.Icon{Value: "icons/update-available.png"}
	iconUpdateOK        = &aw.Icon{Value: "icons/update-ok.png"}
	iconVSCode          = &aw.Icon{Value: "icons/vscode.png"}
	iconWarning         = &aw.Icon{Value: "icons/warning.png"}
	spinnerIcons        = []*aw.Icon{
		{Value: "icons/spinner-1.png"},
		{Value: "icons/spinner-2.png"},
		{Value: "icons/spinner-3.png"},
	}
)

func init() {
	aw.IconError = iconError
	aw.IconWarning = iconWarning
}

// iconSpinner returns a "frame" for a spinning icon.
func iconSpinner() *aw.Icon {
	n := wf.Config.GetInt("RELOAD_PROGRESS", 0)
	wf.Var("RELOAD_PROGRESS", fmt.Sprintf("%d", n+1))
	return spinnerIcons[n%len(spinnerIcons)]
}
