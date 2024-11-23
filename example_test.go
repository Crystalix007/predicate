package predicate_test

import (
	"fmt"

	"github.com/crystalix007/predicate"
)

func ExampleEvaluate() {
	// A return value of 0 means that the predicate is true, any other value
	// means that it is false.
	predicateProgram := `
		import "net/url"

		u, err := url.Parse(arg0)
		if err != nil {
			return false
		}

		return u.Scheme == "https"
	`

	res, err := predicate.Evaluate(predicateProgram, "https://example.com")
	if err != nil {
		fmt.Printf("error evaluating predicate: %v\n", err)
	}

	fmt.Printf("predicate result: %t\n", res)

	// Output: predicate result: true
}
