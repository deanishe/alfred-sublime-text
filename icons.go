// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package main

import (
	"fmt"

	aw "github.com/deanishe/awgo"
)

// Workflow icons
var (
	iconHelp            = &aw.Icon{Value: "icons/help.png"}
	iconIssue           = &aw.Icon{Value: "icons/issue.png"}
	iconReload          = &aw.Icon{Value: "icons/reload.png"}
	iconSettings        = &aw.Icon{Value: "icons/settings.png"}
	iconURL             = &aw.Icon{Value: "icons/url.png"}
	iconUpdateAvailable = &aw.Icon{Value: "icons/update-available.png"}
	iconUpdateOK        = &aw.Icon{Value: "icons/update-ok.png"}
	iconWarning         = &aw.Icon{Value: "icons/warning.png"}
	iconSublime         = &aw.Icon{Value: "icons/sublime.png"}
	iconVSCode          = &aw.Icon{Value: "icons/vscode.png"}
	// iconDocs            = &aw.Icon{Value: "icons/docs.png"}
	// iconOff             = &aw.Icon{Value: "icons/off.png"}
	// iconOn              = &aw.Icon{Value: "icons/on.png"}
	// iconTrash           = &aw.Icon{Value: "icons/trash.png"}
	spinnerIcons = []*aw.Icon{
		{Value: "icons/spinner-1.png"},
		{Value: "icons/spinner-2.png"},
		{Value: "icons/spinner-3.png"},
	}
)

func init() {
	aw.IconWarning = iconWarning
}

// iconSpinner returns a "frame" for a spinning icon.
func iconSpinner() *aw.Icon {
	n := wf.Config.GetInt("RELOAD_PROGRESS", 0)
	wf.Var("RELOAD_PROGRESS", fmt.Sprintf("%d", n+1))
	return spinnerIcons[n%len(spinnerIcons)]
}
