//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-27
//

package main

// Filterer passes selectively passes through strings
type Filterer func(in <-chan string) <-chan string

// Filter is a chain of Filterers
type Filter struct {
	Funcs []Filterer
}

// Use adds a Filterer to the stack
func (f *Filter) Use(fn Filterer) {
	f.Funcs = append(f.Funcs, fn)
}

// Apply runs the filter on a channel.
func (f *Filter) Apply(in <-chan string) <-chan string {

	var out <-chan string

	// Make stack of handlers
	out = f.Funcs[len(f.Funcs)-1](in)
	for i := len(f.Funcs) - 2; i >= 0; i-- {
		out = f.Funcs[i](out)
	}

	return out
}
