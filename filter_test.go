//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-01-27
//

package main

import (
	"path/filepath"
	"testing"
)

func TestFilter(t *testing.T) {
	data := []struct {
		in, out []string
	}{
		{[]string{""}, []string{}},
		{[]string{"file", "file.txt"}, []string{"file.txt"}},
		{[]string{"file.txt", "file.pdf"}, []string{"file.txt", "file.pdf"}},
		{[]string{"file.mp4", "file.pdf"}, []string{"file.pdf"}},
	}

	for _, td := range data {

		var in = make(chan string)

		// Generate input
		go func(c chan string, data []string) {

			for _, s := range data {
				if s == "" {
					continue
				}
				x := filepath.Ext(s)
				if x == ".mp4" || x == "" {
					continue
				}
				c <- s
			}
			close(in)
		}(in, td.in)

		f := Filter{}
		f.Use(func(in <-chan string) <-chan string {
			var out = make(chan string)

			go func() {
				defer close(out)

				for s := range in {
					out <- s
				}

			}()

			return out
		})

		out := f.Apply(in)
		res := []string{}
		for s := range out {
			res = append(res, s)
		}

		if !strSlicesEqual(res, td.out) {
			t.Errorf("Bad Filter. Expected=%#v, Got=%#v", td.out, res)
		}
	}

}

func strSlicesEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i, s := range s1 {
		if s != s2[i] {
			return false
		}
	}

	return true
}
