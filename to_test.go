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

type FooStruct struct{}

func (f FooStruct) String() string { return "FOO" }

type HumanFoo struct{}

func (f HumanFoo) Human() string { return "a friendly foo" }

func FooFunc(a any) {}

func ExampleHuman() {
	fmt.Println(to.Human('r'))                    // not number
	fmt.Println(to.Human("string💢good"))          // unescaped
	fmt.Println(to.Human(new(FooStruct)))         // has String()
	fmt.Println(to.Human(new(HumanFoo)))          // has Human()
	fmt.Println(to.Human([]rune{'r', 's', 'm'}))  // commas
	fmt.Println(to.Human([]string{"foo", "bar"})) // also commas
	fmt.Println(to.Human(func() {}))              // func1
	fmt.Println(to.Human(FooFunc))                // FooFunc
	// Output:
	// 'r'
	// "string💢good"
	// FOO
	// a friendly foo
	// ['r','s','m']
	// ["foo","bar"]
	// func1
	// FooFunc
}

func ExampleDedent_simple() {
	fmt.Printf("%q\n", to.Dedented("\n    foo\n    bar"))
	// Output:
	// "foo\nbar"
}

func ExampleDedent_multiple_Blank_Lines() {
	fmt.Printf("%q\n", to.Dedented("\n\n   \n\n    foo\n    bar"))
	fmt.Printf("%q\n", to.Dedented("\n   \n\n  \n   some"))
	// Output:
	// "foo\nbar"
	// "some"
}

func ExampleDedent_accidental_Chop() {
	fmt.Printf("%q\n", to.Dedented("\n\n   \n\n    foo\n   bar"))
	// Output:
	// "foo\nar"
}

func ExampleIndentation() {
	fmt.Println(to.Indentation("    some"))
	fmt.Println(to.Indentation("  some"))
	fmt.Println(to.Indentation("some"))
	fmt.Println(to.Indentation(" 💚ome"))
	// Output:
	// 4
	// 2
	// 0
	// 1
}

func ExampleHardWrapped() {
	fmt.Printf("%q\n", to.HardWrapped("some thing", 3))
	fmt.Printf("%v\n", to.HardWrapped("There I was not knowing what to do about this exceedingly long line and knowing that certain people would shun me for injecting\nreturns wherever I wanted.", 40))
	// Output:
	// "some\nthing"
	// There I was not knowing what to do about
	// this exceedingly long line and knowing
	// that certain people would shun me for
	// injecting
	// returns wherever I wanted.
}
