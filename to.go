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
	"regexp"
	"runtime"
	"strings"
	"unicode"

	"github.com/rwxrob/fn/filt"
	"github.com/rwxrob/fn/maps"
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

// IndentWrapped adds the specified number of spaces to the beginning of
// every line ensuring that the wrapping is preserved to the specified
// width. Carriage returns (if any) are dropped.
func IndentWrapped(in string, indent, width int) string {
	return Prefixed(Wrapped(in, width-indent), strings.Repeat(" ", indent))
}

// Prefixed returns a string where every line is prefixed. Carriage
// returns (if any) are dropped.
func Prefixed(in, pre string) string {
	lines := Lines(in)
	lines = maps.Prefix(lines, pre)
	return strings.Join(lines, "\n")
}

// Dedented discards any initial lines with nothing but spaces in them and
// then detects the number of space characters at the beginning of the
// first line to the first non-space rune and then subsequently removes
// exactly that many of runes from every following line treating empty
// lines as if they had only n number of spaces. Note that if any line
// does not have n number of initial spaces it the initial runes will
// still be removed. It is, therefore, up to the content creator to
// ensure that all lines have the same space indentation.
func Dedented(in string) string {
	isblank := regexp.MustCompile(`^\s*$`)
	lines := Lines(in)
	var n int
	for {
		if len(lines[n]) == 0 || isblank.MatchString(lines[n]) {
			n++
			continue
		}
		break
	}
	starts := n
	indent := Indentation(lines[n])
	for ; n < len(lines); n++ {
		lines[n] = lines[n][indent:]
	}
	return strings.Join(lines[starts:], "\n")
}

// Indentation returns the number of spaces (in bytes) between beginning
// of the passed string and the first non-space rune.
func Indentation(in string) int {
	var n int
	var v rune
	for n, v = range in {
		if v != ' ' {
			break
		}
	}
	return n
}

// peekWord returns the runes up to the next space.
func peekWord(buf []rune, start int) []rune {
	word := []rune{}
	for _, r := range buf[start:] {
		if unicode.IsSpace(r) {
			break
		}
		word = append(word, r)
	}
	return word
}

// Wrapped expects a string optionally containing line returns (\n)
// that will be kept as hard wrap line boundaries and returns every
// other line exceeding the specified width as one or more wrapped
// lines. All spaces are crunched into a single space. If passed
// a negative width effectively joins all words in the buffer into
// a single line with no wrapping.
func Wrapped(buf string, width int) string {
	if width == 0 {
		return buf
	}
	nbuf := ""
	curwidth := 0
	for i, r := range []rune(buf) {
		// hard breaks always as is
		if r == '\n' {
			nbuf += "\n"
			curwidth = 0
			continue
		}
		if unicode.IsSpace(r) {
			// FIXME: don't peek every word, only after passed width
			// change the space to a '\n' in the buffer slice directly
			next := peekWord([]rune(buf), i+1)
			if width > 0 && (curwidth+len(next)+1) > width {
				nbuf += "\n"
				curwidth = 0
				continue
			}
		}
		nbuf += string(r)
		curwidth++
	}
	return nbuf
}

// UsageGroup joins the slice with bars (|) and wraps with parentheses
// suitable for listing as a group within most command usage strings.
// Empty args are ignored and if no args are passed returns empty
// string. If only one arg, then return just that same arg. Note that no
// transformation is done to the string itself (such as removing white
// space).
func UsageGroup(args []string) string {
	args = filt.NotEmpty(args)
	switch len(args) {
	case 0:
		return ""
	case 1:
		return args[0]
	default:
		return "(" + strings.Join(args, "|") + ")"
	}
}
