# Predicate

Runtime predicate evaluation for Go.

## Example

```go
package main

import (
	"fmt"

	"github.com/Crystalix007/predicate"
)

const indentedPredicateProgram = `
    import "strings"

    return strings.HasPrefix(arg0, "\t")
`

func main() {
	res, err := predicate.Evaluate(indentedPredicateProgram, "some text")
	if err != nil {
		fmt.Printf("error evaluating predicate: %v\n", err)
	}

	fmt.Printf("1st predicate result: %t\n", res)

	res, err = predicate.Evaluate(indentedPredicateProgram, "\tindented text")
	if err != nil {
		fmt.Printf("error evaluating predicate: %v\n", err)
	}

	fmt.Printf("2nd predicate result: %t\n", res)

	// Output: 1st predicate result: false
	// 2nd predicate result: true
}
```
