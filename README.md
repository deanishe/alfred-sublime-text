
<div align="center">
    <img src="icon.png" width="128" height="128">
</div>

Sublime Text Projects Alfred Workflow
=====================================

View, filter and open your Sublime Text (or VSCode) project files.

![][demo]

<!-- MarkdownTOC autolink="true" autoanchor="true" -->

- [Download & Installation](#download--installation)
    - [Catalina and later](#catalina-and-later)
- [Usage](#usage)
    - [Universal Actions](#universal-actions)
    - [Hotkeys](#hotkeys)
    - [External Triggers](#external-triggers)
- [How it works](#how-it-works)
- [Configuration](#configuration)
- [Licensing, thanks](#licensing-thanks)

<!-- /MarkdownTOC -->


<a id="download--installation"></a>
Download & Installation
-----------------------

Download the workflow from [GitHub][gh-releases] and install by double-clicking the `Sublime-Text-Projects-X.X.X.alfredworkflow` file.


<a id="catalina-and-later"></a>
### Catalina and later

If you're running Catalina or later (macOS 10.15+), you'll need to [grant the workflow executable permission to run][catalina].


<a id="usage"></a>
Usage
-----

There is one keyword, `.st`, which works as follows:

- `.st [<query>]` — List/filter your `.sublime-project` files
	+ `↩` — Open result in Sublime Text
	+ `⌘+↩` — Reveal file in Finder
- `.st rescan` — Reload cached list of projects
- `.st config` — Show the current settings
    - `Workflow Is Up To Date` / `Workflow Update Available` — Install update or check for update
    - `Rescan Projects` — Reload list of projects
    - `Edit Config File` — Open workflow's configuration file
    - `Editor: Sublime Text` / `Editor: VS Code` — Which editor is selected
    - `Action Project File` — Whether copying/actioning a search result should use the path of the project file instead of that of the first project directory
    - `View Help File` — Open README in your browser
    - `Report Issue` — Open GitHub issue tracker in your browser
    - `Visit Forum Thread` — Open workflow's thread on [alfredforum.com][forum]

You can enter `search` or `config` as a search query anywhere to jump to the corresponding screen.


<a id="universal-actions"></a>
### Universal Actions

There are Universal Actions for files, URLs and text. Files are opened, and text is inserted into a new document.

Multiple URLs are treated as text, but a single URL is retrieved with curl and a new document is created with its contents.


<a id="hotkeys"></a>
### Hotkeys

The workflow has two Hotkeys (marked red) that you can set to open the currently-selected files in any application in Sublime Text. One Hotkey is for Finder and Path Finder only, and the other is for all other applications. You should set them to the same keyboard shortcut.

The Finder/Path Finder variant doesn't rely on Alfred's "Selection in macOS" feature, and will open the frontmost window's target (the folder whose contents it's showing) if nothing is selected.


<a id="external-triggers"></a>
### External Triggers

The workflow has the following External Triggers that can be used from scripts or other workflows:

|    Name    |                      Description                       |
|------------|--------------------------------------------------------|
| `new`      | Create a new document containing the given text        |
| `open`     | Open the specified path in Sublime Text                |
| `open-url` | Create a new document with the data retrieved from URL |
| `search`   | Show project search results for given query            |


<a id="how-it-works"></a>
How it works
------------

The workflow scans your system for `.sublime-project` (or `.code-workspace`) files using `locate`, `mdfind` and (optionally) `find`. It then caches the list of projects for 10 minutes (by default).

As the `locate` database isn't enabled on most machines (and isn't updated frequently in any case), and `mdfind` ignores hidden directories, there is an additional, optional `find`-based scanner to "fill the gaps", which you must specifically configure (see below).

**NOTE**: When the workflow is asked to open a directory (e.g. via External Trigger or Universal Action), it looks for a project file in the directory, and opens that instead if one is found.


<a id="configuration"></a>
Configuration
-------------

Scan intervals are configured in the [workflow's configuration sheet in Alfred Preferences][confsheet]:

|        Variable       |   Type   |                          Usage                           |
|-----------------------|----------|----------------------------------------------------------|
| `INTERVAL_FIND`       | `duration` | How long to cache `find` search results for              |
| `INTERVAL_LOCATE`     | `duration` | How long to cache `locate` search results for            |
| `INTERVAL_MDFIND`     | `duration` | How long to cache `mdfind` search results for            |
| `ACTION_PROJECT_FILE` | `boolean`  | Copying/actioning a search result uses project file path |
| `VSCODE`              | `boolean`  | Switch to Visual Studio Code mode                        |

`duration` values should be of the form `10m` or `2h`. Set to `0` to disable a particular scanner.
`boolean` values should be of the form `true` and `false` or `1` and `0`.

The workflow should work "out of the box", but if you have project files in directories that `mdfind` doesn't see (hidden directories, network shares), you may have to explicitly add some search paths to the `sublime.toml` configuration file in the workflow's data directory. The file is created on first run, and you can use `.st config > Workflow Settings > Edit Config File` to open it.

These directories are searched with `find`.

You can also add glob patterns to the `excludes` list in the settings file to ignore certain results. Excludes apply to all scanners.

The options are documented in the settings file itself.


<a id="licensing-thanks"></a>
Licensing, thanks
-----------------

All the code is released under the [MIT Licence][mit].

The workflow is based on the [AwGo workflow library][awgo], also released under the [MIT Licence][mit].

The icons are based on [Font Awesome][awesome] and [Material Design Icons][matcom].

[forum]: https://www.alfredforum.com
[awgo]: https://github.com/deanishe/awgo
[awesome]: https://fontawesome.com
[matcom]: https://materialdesignicons.com/
[demo]: https://raw.githubusercontent.com/deanishe/alfred-sublime-text/master/demo.gif
[gh-releases]: https://github.com/deanishe/alfred-sublime-text/releases/latest
[mit]: http://opensource.org/licenses/MIT
[confsheet]: https://www.alfredapp.com/help/workflows/advanced/variables/#environment
[catalina]: https://github.com/deanishe/awgo/wiki/Catalina
