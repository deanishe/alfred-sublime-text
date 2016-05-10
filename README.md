# Sublime Text Projects Alfred Workflow #

View, filter and open your Sublime Text (2 and 3) project files.

![][demo]

## Installation ##

Download the workflow from [GitHub][gh-releases] or [Packal][packal].

## Usage ##

- `.st [<query>]` — List/filter your `.sublime-project` files
	+ `↩` — Open result in Sublime Text
	+ `⌘+↩` — Reveal file in Finder
- `.stconfig` — Show the current settings
- `.sthelp` — View the included help file

**Note**: You can currently only alter the settings by editing the `settings.json` file by hand. Hit `↩` on the **Edit Configuration** item to open it in your default JSON editor.

## Licensing, thanks ##

All the code is released under the [MIT Licence][mit].

The workflow is based on the [Alfred-Workflow library][alfred-workflow], also released under the [MIT Licence][mit].

The icons are by [dmatarazzo][dmatarazzo].


[alfred-workflow]: http://www.deanishe.net/alfred-workflow/
[demo]: https://raw.githubusercontent.com/deanishe/alfred-sublime-text/master/demo.gif
[gh-releases]: https://github.com/deanishe/alfred-sublime-text/releases/latest
[packal]: http://www.packal.org/workflow/sublime-text-projects
[mit]: http://opensource.org/licenses/MIT
[dmatarazzo]: https://github.com/dmatarazzo/Sublime-Text-2-Icon
