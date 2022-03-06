package to_test

import (
	"fmt"

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
