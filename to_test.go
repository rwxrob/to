// Copyright 2022 Robert S. Muhlestein
// SPDX-License-Identifier: Apache-2.0

package to_test

import (
	"fmt"

	"github.com/rwxrob/fn"
	"github.com/rwxrob/fn/each"
	"github.com/rwxrob/to"
)

type stringer struct{}

func (s stringer) String() string { return "stringer" }

func ExampleString() {
	stuff := []any{
		"some", []byte{'t', 'h', 'i', 'n', 'g'},
		1, 2.234, stringer{},
	}
	for _, s := range stuff {
		fmt.Printf("%q ", to.String(s))
	}
	// Output:
	// "some" "thing" "1" "2.234" "stringer"

}

func Foo() {}

func ExampleFuncName() {

	f1 := func() {}
	f2 := func() {}

	// Foo is defined outside of the ExampleFuncName

	each.Println(fn.Map([]any{f1, f2, Foo, to.Lines}, to.FuncName))

	// Output:
	// func1
	// func2
	// Foo
	// Lines
}

func ExampleLines() {
	buf := `
some

thing 
here

mkay
`
	each.Print(to.Lines(buf))
	// Output:
	// something heremkay
}
