#!/usr/bin/osascript -l JavaScript

ObjC.import('stdlib')

// Application bundle IDs
const finderId = 'com.apple.Finder',
	pathFinderId = 'com.cocoatech.PathFinder'

// Get environment variable
function getEnv(key) {
	try {
		return $.getenv(key)
	} catch(e) {
		return null
	}
}

// Return Path Finder selection or target as POSIX paths
function pathFinderPaths() {
	const pf = Application(pathFinderId)
	let selection = pf.selection()
	// selected files
	if (selection) return selection.map(pfi => pfi.posixPath())
	// target of frontmost window
	return [pf.finderWindows[0].target.posixPath()]
}

// Return Finder selection or target as POSIX paths
function finderPaths() {
	const file2Path = fi => Path(decodeURI(fi.url()).slice(7)).toString()
	const finder = Application(finderId)
	let selection = finder.selection()
	// selected files
	if (selection && selection.length) return selection.map(file2Path)
	// target of frontmost window
	return [file2Path(finder.finderWindows[0].target)]
}

function run() {
	const activeApp = getEnv('focusedapp')
	let paths = []
	console.log(`🍻\nactiveApp=${activeApp}`)

	if (activeApp === pathFinderId) paths = pathFinderPaths()
	else paths = finderPaths()

	return JSON.stringify({alfredworkflow: {arg: paths}})
}