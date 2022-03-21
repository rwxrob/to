// Copyright 2022 Robert S. Muhlestein
// SPDX-License-Identifier: Apache-2.0

/*

Package to contains a number of converters that take any number of types and return something transformed from them. It also contains a more granular approach to fmt.Stringer.

*/
package to

import (
	"bufio"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// Stringer interfaces fulfills fmt.Stringer with the additional promise
// that the output of String method will always be both consistently
// parsable (say as JSON) and will never span more than a single line.
// Stringer also requires the StringLong method promising to produce
// parsable strings that span multiple lines to remain easy to read.
type Stringer interface {
	String() string
	StringLong() string
}

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

// HumanFriend implementations have a human readable form that is even
// friendlier than fmt.Stringer.
type HumanFriend interface {
	Human() string
}

// Human returns a human-friendly string version of the item,
// specifically:
//
//     * single-quoted runes
//     * double-quoted strings
//     * numbers as numbers
//     * function names are looked up
//     * slices joined with "," and wrapped in []
//
// Anything else is rendered as its fmt.Sprintf("%v",it) form.
func Human(a any) string {
	switch v := a.(type) {

	case string:
		return fmt.Sprintf("%q", v)

	case rune:
		return fmt.Sprintf("%q", v)

	case []string:
		st := []string{}
		for _, r := range v {
			st = append(st, fmt.Sprintf("%q", r))
		}
		return "[" + strings.Join(st, ",") + "]"

	case []rune:
		st := []string{}
		for _, r := range v {
			st = append(st, fmt.Sprintf("%q", r))
		}
		return "[" + strings.Join(st, ",") + "]"

	case []any:
		st := []string{}
		for _, r := range v {
			st = append(st, Human(r))
		}
		return "[" + strings.Join(st, ",") + "]"

	case HumanFriend:
		return v.Human()

	default:
		typ := fmt.Sprintf("%v", reflect.TypeOf(a))
		if len(typ) > 3 && typ[0:4] == "func" {
			return FuncName(a)
		}
		return fmt.Sprintf("%v", a)

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
