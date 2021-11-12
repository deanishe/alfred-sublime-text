#!/usr/bin/osascript -l JavaScript

ObjC.import('stdlib')

const finderId = 'com.apple.Finder',
	pathFinderId = 'com.cocoatech.PathFinder'

const file2Path = fi => Path(decodeURI(fi.url()).slice(7)).toString()

function getEnv(key) {
	try {
		return $.getenv(key)
	} catch(e) {
		return null
	}
}

function pathFinderPaths() {
	const pf = Application(pathFinderId)
	let selection = pf.selection(),
		paths = []

	if (selection) {
		selection.forEach(pfi => {
			let p = pfi.posixPath()
			console.log(`[Path Finder] selection=${p}`)
			paths.push(p)
		})
	} else {
		let p = pf.finderWindows[0].target.posixPath()
		console.log(`[Path Finder] target=${p}`)
		paths.push(p)
	}

	return paths
}

function finderPaths() {
	const finder = Application(finderId)
	let paths = [],
		selection = finder.selection()

	if (selection) {
		selection.forEach(fi => {
			let p = file2Path(fi)
			console.log(`[Finder] selection=${p}`)
			paths.push(p)
		})
	} else {
		let p = file2Path(finder.finderWindows[0].target)
		console.log(`[Finder] target=${p}`)
		paths.push(p)
	}

	return paths
}

function run() {
	console.log('🍻')
	const activeApp = getEnv('focusedapp')
	console.log(`activeApp=${activeApp}`)
	let paths = []
	if (activeApp === pathFinderId) {
		paths = pathFinderPaths()
	} else {
		paths = finderPaths()
	}

	return JSON.stringify({alfredworkflow: {arg: paths}})
}