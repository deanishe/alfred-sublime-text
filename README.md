
Sublime Text Projects Alfred Workflow
=====================================

View, filter and open your Sublime Text 3 project files.

![][demo]


Download & Installation
-----------------------

Download the workflow from [GitHub][gh-releases] and install by double-clicking the `Sublime-Text-Projects-X.X.X.alfredworkflow` file.


Usage
-----

- `.st [<query>]` — List/filter your `.sublime-project` files
	+ `↩` — Open result in Sublime Text
	+ `⌘+↩` — Reveal file in Finder
- `.stconfig` — Show the current settings
    - `Workflow Is Up To Date` / `Workflow Update Available` — Install update or check for update
    - `Edit Config File` — Open workflow's config file in Sublime Text
    - `View Help File` — Open README in your browser
    - `Report Issue` — Open GitHub issue tracker in your browser
    - `Visit Forum Thread` — Open workflow's thread on [alfredforum.com][forum]

**Note**: You must edit the configuration and add some directories to search before using the workflow, or it won't do anything.


Configuration
-------------

The workflow is configured by editing the `sublime.toml` file in the workflow's data directory. It will be created by the workflow on first run, and you can use `.stconfig > Edit Config File` to open it in Sublime Text.

The available options are documented in the settings file itself.


Licensing, thanks
-----------------

All the code is released under the [MIT Licence][mit].

The workflow is based on the [AwGo library][awgo], also released under the [MIT Licence][mit].

The icons are based on [Font Awesome][awesome] and [Material Design Icons][matcom].

[forum]: https://www.alfredforum.com
[awgo]: https://github.com/deanishe/awgo
[awesome]: https://fontawesome.com
[matcom]: https://materialdesignicons.com/
[demo]: https://raw.githubusercontent.com/deanishe/alfred-sublime-text/master/demo.gif
[gh-releases]: https://github.com/deanishe/alfred-sublime-text/releases/latest
[mit]: http://opensource.org/licenses/MIT

