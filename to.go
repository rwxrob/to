// Copyright 2022 Robert S. Muhlestein
// SPDX-License-Identifier: Apache-2.0

/*

Package to contains a number of converters that take any number of types and return something transformed from them.

*/
package to

import (
	"bufio"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// String converts whatever is passed to its fmt.Sprintf("%v") string
// version (but avoids calling it if possible). Be sure you use things
// with consistent string representations.
func String(in any) string {
	switch v := in.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case []rune:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// FuncName makes a best effort attempt to return the string name of the
// passed function. Anonymous functions are named "funcN" where N is the
// order of appearance within the current scope. Note that this function
// will panic if not passed a function.
func FuncName(i any) string {
	p := runtime.FuncForPC(reflect.ValueOf(i).Pointer())
	n := strings.Split(p.Name(), `.`)
	return n[len(n)-1]
}

// Lines transforms the input into a string and then divides that string
// up into lines (\r?\n) suitable for functional map operations.
func Lines(in any) []string {
	buf := String(in)
	lines := []string{}
	s := bufio.NewScanner(strings.NewReader(buf))
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	return lines
}
