// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/deanishe/awgo/util/build"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

const (
	buildDir = "./build"
	distDir  = "./dist"
)

var (
	info       *build.Info
	workDir    string
	archs      = []string{"amd64", "arm64"}
	codeSignID = os.Getenv("CODESIGN_ID")
)

func init() {
	var err error
	if info, err = build.NewInfo(); err != nil {
		panic(err)
	}
	if workDir, err = os.Getwd(); err != nil {
		panic(err)
	}
}

func mod(args ...string) error {
	argv := append([]string{"mod"}, args...)
	return sh.RunWith(info.Env(), "go", argv...)
}

// Aliases are mage command aliases.
var Aliases = map[string]interface{}{
	"b": Build,
	"c": Clean,
	"d": Dist,
	"l": Link,
}

// make workflow in build directory
func Build() error {
	mg.Deps(cleanBuild)
	fmt.Println("building ...")

	var bins []string
	for _, arch := range archs {
		env := info.Env()
		env["GOOS"] = "darwin"
		env["GOARCH"] = arch
		bin := fmt.Sprintf("%s/alfred-sublime.%s", buildDir, arch)
		bins = append(bins, bin)
		if err := sh.RunWith(env, "go", "build", "-o", bin, "."); err != nil {
			return err
		}
	}

	// build fat binary
	args := []string{"-create", "-output", filepath.Join(buildDir, "alfred-sublime")}
	if err := sh.Run("/usr/bin/lipo", append(args, bins...)...); err != nil {
		return err
	}
	// delete arch-specific binaries
	for _, bin := range bins {
		if err := sh.Rm(bin); err != nil {
			return err
		}
	}

	// files to include in workflow
	globs := build.Globs(
		"*.js",
		"*.png",
		"info.plist",
		"*.html",
		"README.md",
		"LICENCE.txt",
		"icons/*.png",
	)

	return build.SymlinkGlobs(buildDir, globs...)
}

// run workflow
func Run() error {
	mg.Deps(Build)
	fmt.Println("running ...")
	if err := os.Chdir("./build"); err != nil {
		return err
	}
	defer os.Chdir(workDir)

	return sh.RunWith(info.Env(), "./alfred-sublime", "-h")
}

func codeSign() error {
	if codeSignID == "" {
		fmt.Println("skipping signing: CODESIGN_ID unset")
		return nil
	}
	return sh.Run("codesign", "-f", "-s", codeSignID, "-i", info.BundleID, filepath.Join(buildDir, "alfred-sublime"))
}

// create an .alfredworkflow file in ./dist
func Dist() error {
	mg.SerialDeps(Clean, Build, codeSign)
	p, err := build.Export(buildDir, distDir)
	if err != nil {
		return err
	}

	fmt.Printf("built workflow file %s\n", p)
	return nil
}

// symlink build directory to Alfred's workflow directory
func Link() error {
	mg.Deps(Build)

	fmt.Printf("linking %s to workflow directory ...\n", buildDir)
	target := filepath.Join(info.AlfredWorkflowDir, info.BundleID)

	if exists(target) {
		fmt.Println("removing existing workflow ...")
	}
	// try to remove it anyway, as dangling symlinks register as existing
	if err := os.RemoveAll(target); err != nil && !os.IsNotExist(err) {
		return err
	}

	src, err := filepath.Abs(buildDir)
	if err != nil {
		return err
	}
	return build.Symlink(target, src, true)
}

// download dependencies
func Deps() error {
	mg.Deps(cleanDeps)
	fmt.Println("downloading deps ...")
	return mod("download")
}

func cleanDeps() error { return mod("tidy", "-v") }

// remove build files
func Clean() { mg.Deps(cleanBuild, cleanMage) }

func cleanBuild() error {
	fmt.Printf("cleaning %s ...\n", buildDir)
	if err := sh.Rm(buildDir); err != nil {
		return err
	}
	return os.MkdirAll(buildDir, 0755)
}

func cleanMage() error {
	fmt.Println("cleaning mage ...")
	return sh.Run("mage", "-clean")
}

// return true if path exists
func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}

	return true
}
