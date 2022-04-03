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

	"github.com/rwxrob/fn/maps"
	"github.com/rwxrob/structs/qstack"
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
// width. See Wrapped.
func IndentWrapped(in string, indent, width int) string {
	wwidth := width - indent
	body, _ := Wrapped(in, wwidth)
	var buf string
	for _, line := range Lines(body) {
		buf += fmt.Sprintln(strings.Repeat(" ", indent) + line)
	}
	return buf
}

// Prefixed returns a string where every line is prefixed. Carriage
// returns (if any) are dropped.
func Prefixed(in, pre string) string {
	lines := Lines(in)
	lines = maps.Prefix(lines, pre)
	return strings.Join(lines, "\n")
}

// Dedented discards any initial lines with nothing but whitespace in
// them and then detects the number and type of whitespace characters at
// the beginning of the first line to the first non-whitespace rune and
// then subsequently removes that number of runes from every
// following line treating empty lines as if they had only n number of
// spaces.  Note that if any line does not have n number of initial
// spaces it the initial runes will still be removed. It is, therefore,
// up to the content creator to ensure that all lines have the same
// space indentation.
func Dedented(in string) string {
	isblank := regexp.MustCompile(`^\s*$`)
	lines := Lines(in)
	var n int
	for len(lines[n]) == 0 || isblank.MatchString(lines[n]) {
		n++
	}
	starts := n
	indent := Indentation(lines[n])
	for ; n < len(lines); n++ {
		if len(lines[n]) > indent {
			lines[n] = lines[n][indent:]
		}
	}
	return strings.Join(lines[starts:], "\n")
}

// Indentation returns the number of whitespace runes (in bytes) between
// beginning of the passed string and the first non-whitespace rune.
func Indentation(in string) int {
	var n int
	var v rune
	for n, v = range in {
		if !unicode.IsSpace(v) {
			break
		}
	}
	return n
}

// Wrapped will return a word wrapped string at the given boundary width
// (in bytes) and the count of words contained in the string.  All
// whitespace is compressed to a single space. Any width less than
// 1 will simply trim and crunch whitespace returning essentially the
// same string and the word count.  If the width is less than any given
// word at the start of a line than it will be the only word on the line
// even if the word length exceeds the width. Non attempt at
// word-hyphenation is made. Note that whitespace is defined as
// unicode.IsSpace and does not include control characters.
func Wrapped(it string, width int) (string, int) {
	words := qstack.Fields(it)
	if width < 1 {
		return strings.Join(words.Items(), " "), words.Len
	}
	var curwidth int
	var wrapped string
	var line []string
	for words.Scan() {
		cur := words.Current()
		if len(line) == 0 {
			line = append(line, cur)
			curwidth += len(cur) + 1
			continue
		}
		if curwidth+len(cur) > width {
			wrapped += strings.Join(line, " ") + "\n"
			curwidth = len(cur) + 1
			line = []string{cur}
			continue
		}
		line = append(line, cur)
		curwidth += len(cur) + 1
	}
	wrapped += strings.Join(line, " ")
	return wrapped, words.Len
}
