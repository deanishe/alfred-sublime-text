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
	green = "03ae03"
	// blue  = "5484f3"
	// red    = "b00000"
	// yellow = "f8ac30"
)

const (
	buildDir = "./build"
	distDir  = "./dist"
)

var (
	info     *build.Info
	workDir string
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
	mg.Deps(cleanBuild, Icons)
	fmt.Println("building ...")

	if err := sh.RunWith(info.Env(), "go", "build", "-o", "./build/alfsubl", "."); err != nil {
		return err
	}

	// files to include in workflow
	globs := build.Globs(
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

	return sh.RunWith(info.Env(), "./alfsubl", "-h")
}

// create an .alfredworkflow file in ./dist
func Dist() error {
	mg.SerialDeps(Clean, Build)
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

// generate icons
func Icons() error {

	copies := []struct {
		src, dest, colour string
	}{
		{"docs.png", "help.png", green},
	}

	for i, cfg := range copies {

		src := filepath.Join("icons", cfg.src)
		dest := filepath.Join("icons", cfg.dest)

		if exists(dest) {
			fmt.Printf("[%d/%d] skipped existing: %s\n", i+1, len(copies), dest)
			continue
		}

		if err := copyImage(src, dest, cfg.colour); err != nil {
			return err
		}
	}

	return rotateIcon("./icons/loading.png", []int{15, 30})
}

// download dependencies
func Deps() error {
	mg.Deps(cleanDeps)
	fmt.Println("downloading deps ...")
	return mod("download")
}

func cleanDeps() error {  return mod("tidy", "-v") }

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