title: Sublime Text Projects for Alfred 2 Help

# Sublime Text Projects for Alfred 2 #

View, filter and open Sublime Text (2 and 3) projects files saved on your Mac.

![](https://raw.githubusercontent.com/deanishe/alfred-sublime-text/master/demo.gif "")

- [Usage](#usage)
- [How it works](#howitworks)
	+ [Excludes](#excludes)
	+ [find](#find)
	+ [locate](#locate)
		* [Turning on locate](#turningonlocate)
		* [Forcing an update](#forcinganupdate)
- [Settings](#settings)

## Usage ##

- `.st [<query>]` — List/filter your `.sublime-project` files
	+ `↩` — Open result in Sublime Text
	+ `⌘+↩` — Reveal file in Finder
- `.stconfig` — Show the current settings
- `.sthelp` — View this help file

**Note**: You can currently only alter the settings by editing the `settings.json` file by hand. Hit `↩` on the **Edit Configuration** item to open it in your default JSON editor.

## How it works ##

By default, the workflow uses `mdfind` (which in turn uses Spotlight's index) to find all the `.sublime-project` files on your system.

Unfortunately, `mdfind` won't find hidden files or files in hidden directories. Thus, you can also add additional search directories, which will be searched for `.sublime-project` files using `find`, and/or turn on the `locate` database.

**Note**: Although it can take several seconds to perform the search with `find` and/or `locate`, the results are cached and the cache is updated in the background, so the Workflow will always remain responsive, although it might take a few seconds for newly-added files to show up in the results.

### Excludes ###

As `locate` in particular will likely return a lot of results you don't want to see, such as `.sublime-project` files that are part of apps' bundles, you can also add globbing patterns to the `settings.json` file, and paths that match these patterns will be ignored.

See [Settings](#settings) for more details.

### find ###

`find` is a common UNIX command for recursively searching directories. In contrast to `mdfind` and `locate`, it does not use a pre-compiled database, so it shouldn't be used on large directory hierarchies. `locate` (see [below](#locate)) is a better option for such directories. Unfortunately, the `locate` database is updated infrequently, so you might want to consider [forcing an update](#forcinganupdate).

By default, the directories `~/.config` and `~/.dotfiles` are added to the configuration, but will be ignored if they don't exist.

### locate ###

`locate` is a common UNIX command that maintains a list of all files on your system to enable (relatively) fast searching.

By default, its database is only updated once a week (on Saturdays at 3.15 a.m. on OSX). It is *a lot* slower than `mdfind` (which uses the Spotlight index), but also indexes hidden files.

So, if you have Sublime Project files stored in hidden directories, you might want to consider turning on `locate`.

#### Turning on locate ####

Execute the following command in `Terminal` to activate `locate`:

```
sudo launchctl load -w /System/Library/LaunchDaemons/com.apple.locate.plist
```

#### Forcing an update ####

Execute the following command in `Terminal` to force an update of the `locate` database:

```
sudo /usr/libexec/locate.updatedb
```

## Settings ##

The workflow configuration is stored in `setttings.json` in the Workflow's data directory. It's assumed that as a user of Sublime Text, you know how to edit a JSON file ;)

You can view the settings (and open `settings.json` in your editor) by entering `.stconfig` in Alfred.

The default config file looks like this:

```
{
  "excludes": [
    "/Applications/*.app/*"
  ],
  "locatedb_cached": 0,
  "search_directories": [
    "~/.dotfiles",
    "~/.config"
  ]
}
```

Add [globbing patterns](https://docs.python.org/2/library/fnmatch.html#module-fnmatch) to the `excludes` list to remove them from the results.

Add hidden directories to `search_directories` if you want them to be searched, too.

**Note**: Hidden directories will be searched with `find`, so don't add large directory hierarchies.
