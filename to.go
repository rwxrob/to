// Copyright 2022 Robert S. Muhlestein
// SPDX-License-Identifier: Apache-2.0

/*

Package to contains a number of converters that take any number of types and return something transformed from them. It also contains a more granular approach to fmt.Stringer.

*/
package to

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/rwxrob/fn/maps"
	"github.com/rwxrob/scan"
	"github.com/rwxrob/structs/qstack"
)

type Text interface{ string | []rune }

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

// Bytes converts whatever is passed into a []byte slice. Logs and
// returns nil if it cannot convert. Supports the following types:
// string, []byte, []rune, io.Reader.
func Bytes(in any) []byte {
	switch v := in.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	case []rune:
		return []byte(string(v))
	case io.Reader:
		buf, err := io.ReadAll(v)
		if err != nil {
			log.Println(err)
		}
		return buf
	default:
		log.Printf("cannot convert %T to bytes", in)
		return nil
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

// Indented returns a string with each line indented by the specified
// number of spaces. Carriage returns are stripped (if found) as
// a side-effect.
func Indented(in string, indent int) string {
	var buf string
	for _, line := range Lines(in) {
		buf += fmt.Sprintln(strings.Repeat(" ", indent) + line)
	}
	return buf
}

// IndentWrapped adds the specified number of spaces to the beginning of
// every line ensuring that the wrapping is preserved to the specified
// width. See Wrapped.
func IndentWrapped(in string, indent, width int) string {
	wwidth := width - indent
	body, _ := Wrapped(in, wwidth)
	return Indented(body, indent)
}

// Prefixed returns a string where every line is prefixed. Carriage
// returns (if any) are dropped.
func Prefixed(in, pre string) string {
	lines := Lines(in)
	lines = maps.Prefix(lines, pre)
	return strings.Join(lines, "\n")
}

// Dedented discards any initial blank lines with nothing but whitespace in
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
		if len(lines[n]) >= indent {
			lines[n] = lines[n][indent:]
		}
	}
	return strings.Join(lines[starts:], "\n")
}

// Indentation returns the number of whitespace runes (in bytes) between
// beginning of the passed string and the first non-whitespace rune.
func Indentation[T Text](in T) int {
	var n int
	var v rune
	for n, v = range []rune(in) {
		if !unicode.IsSpace(v) {
			break
		}
	}
	return n
}

// RuneCount returns the actual number of runes of the string only
// counting the unicode.IsGraphic runes. All others are ignored.  This
// is critical when calculating line lengths for terminal output where
// the string contains escape characters. Note that some runes will
// occupy two columns instead of one depending on the terminal.
func RuneCount[T string | []byte | []rune](in T) int {
	var c int
	s := scan.R{B: []byte(string(in))}
	for s.Scan() {
		if unicode.IsGraphic(s.R) {
			c++
		}
	}
	return c
}

// Words will return the string will all contiguous runs of
// unicode.IsSpace runes converted into a single space. All leading and
// trailing white space will also be trimmed.
func Words(it string) string {
	return strings.Join(qstack.Fields(it).Items(), " ")
}

// Wrapped will return a word wrapped string at the given boundary width
// (in bytes) and the count of words contained in the string.  All
// white space is compressed to a single space. Any width less than
// 1 will simply trim and crunch white space returning essentially the
// same string and the word count.  If the width is less than any given
// word at the start of a line than it will be the only word on the line
// even if the word length exceeds the width. Non attempt at
// word-hyphenation is made. Note that white space is defined as
// unicode.IsSpace and does not include control characters. Anything
// that is not unicode.IsSpace or unicode.IsGraphic will be ignored in
// the column count.
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
		count := RuneCount(cur)
		if len(line) == 0 {
			line = append(line, cur)
			curwidth += count
			continue
		}
		if curwidth+count+1 > width {
			wrapped += strings.Join(line, " ") + "\n"
			curwidth = count
			line = []string{cur}
			continue
		}
		line = append(line, cur)
		curwidth += RuneCount(cur) + 1
	}
	wrapped += strings.Join(line, " ")
	return wrapped, words.Len
}

// MergedMaps combines the maps with "last wins" priority. Always
// returns a new map of the given type, even if empty.
func MergedMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	combined := map[K]V{}
	for _, m := range maps {
		for k, v := range m {
			combined[k] = v
		}
	}
	return combined
}

// StopWatch converts a duration into a string that one would expect to
// see on a stopwatch.
func StopWatch(dur time.Duration) string {
	var out string

	sec := dur.Seconds()
	if sec < 0 {
		out += "-"
	}
	sec = math.Abs(sec)

	if sec >= 3600 {
		hours := sec / 3600
		sec = math.Mod(sec, 3600)
		out += fmt.Sprintf("%v:", int(hours))
	}

	if sec >= 60 {
		var form string
		mins := sec / 60
		sec = math.Mod(sec, 60)
		if len(out) == 0 {
			form = `%v:`
		} else {
			form = `%02v:`
		}
		out += fmt.Sprintf(form, int(mins))
	}

	var form string
	if len(out) == 0 {
		form = `%02v`
	} else {
		form = `%02v`
	}
	out += fmt.Sprintf(form, int(sec))

	return out
}

// EscReturns changes any actual carriage returns or line returns into
// their backslashed equivalents and returns a string. This is different
// than Sprintf("%q") since that escapes several other things.
func EscReturns[T string | []byte | []rune](in T) string {
	runes := []rune(string(in))
	var out string
	for _, r := range runes {
		switch r {
		case '\r':
			out += "\\r"
		case '\n':
			out += "\\n"
		default:
			out += string(r)
		}
	}
	return string(out)
}

// UnEscReturns changes any escaped carriage returns or line returns into
// their actual values.
func UnEscReturns[T string | []byte | []rune](in T) string {
	runes := []rune(string(in))
	var out string
	for n := 0; n < len(runes); n++ {
		if runes[n] == '\\' && runes[n+1] == 'r' {
			out += "\r"
			n++
			continue
		}
		if runes[n] == '\\' && runes[n+1] == 'n' {
			out += "\n"
			n++
			continue
		}
		out += string(runes[n])
	}
	return string(out)
}

// HTTPS simply adds the prefix "https://" if not found. Useful for
// allowing non-prefixed URLs and later converting them.
func HTTPS(url string) string {
	if len(url) < 8 || url[0:8] != "https://" {
		return "https://" + url
	}
	return url
}

const IsosecFmt = `20060102150405`

// Isosec converts the passed time into an ISO8601 (RFC3339) string time
// stamp without the T or any punctuation for UTC time.
func Isosec(t time.Time) string {
	return t.UTC().Format(IsosecFmt)
}
