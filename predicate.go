package predicate

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

const programTemplate = `
package predicate

{{.Imports}}

func Do({{range .Args}}
	{{.Name}} {{.Type}},{{end}}) bool {
	{{.Program}}
}
`

type templateArg struct {
	Name string
	Type string
}

type templateData struct {
	Imports string
	Args    []templateArg
	Program string
}

// PanicError is an error that occurs when a panic occurs in the predicate.
type PanicError struct {
	Panic any
}

var (
	// ErrEmptyPredicate is an error that occurs when the predicate is empty.
	ErrEmptyPredicate = errors.New("predicate: empty predicate")

	// ErrPredicateDefinitionInvalid is an error that occurs when the predicate definition is invalid.
	ErrPredicateDefinitionInvalid = errors.New("predicate: definition is invalid")

	// ErrPredicateCallbackReferenceFailed is an error that occurs when the predicate callback reference failed.
	ErrPredicateCallbackReferenceFailed = errors.New("predicate: failed to get reference to callback")

	// ErrPredicateReturnInvalid is an error that occurs when the predicate return value is invalid.
	ErrPredicateReturnInvalid = errors.New("predicate: invalid return value")
)

// Error returns the error message.
func (e *PanicError) Error() string {
	return fmt.Sprintf("predicate: panic while evaluating: %v", e.Panic)
}

// Is checks if the panic error is the same as the target error.
func (e *PanicError) Is(target error) bool {
	if targetPanic, ok := target.(*PanicError); ok {
		return e.Panic == targetPanic.Panic
	}

	if panicErr, ok := e.Panic.(error); ok {
		return errors.Is(panicErr, target)
	}

	return false
}

// Evaluate evaluates the predicate and returns the result.
func Evaluate(predicate string, args ...any) (res bool, err error) {
	interpreter := interp.New(interp.Options{})

	interpreter.Use(stdlib.Symbols)

	importLines, nonImportLines := separateImports(strings.Split(predicate, "\n"))

	if len(nonImportLines) == 0 || !containsReturn(nonImportLines) {
		return false, ErrEmptyPredicate
	}

	bs, err := executeTemplate(importLines, nonImportLines, args...)
	if err != nil {
		return false, err
	}

	_, err = interpreter.Eval(bs.String())
	if err != nil {
		return false, err
	}

	predicateFunc, err := interpreter.Eval("predicate.Do")
	if err != nil {
		return false, err
	}

	if !predicateFunc.CanInterface() {
		return false, ErrPredicateDefinitionInvalid
	}

	defer func() {
		if p := recover(); p != nil {
			res = false
			err = &PanicError{Panic: p}

			if reflectValue, ok := p.(reflect.Value); ok && reflectValue.CanInterface() {
				err = &PanicError{Panic: reflectValue.Interface()}
			}
		}
	}()

	reflectVals := make([]reflect.Value, len(args))

	for i, arg := range args {
		reflectVals[i] = reflect.ValueOf(arg)
	}

	predicateValues := predicateFunc.Call(reflectVals)

	if len(predicateValues) != 1 {
		return false, fmt.Errorf("%w: invalid number of return values", ErrPredicateReturnInvalid)
	}

	if predicateValues[0].Kind() != reflect.Bool {
		return false, fmt.Errorf("%w: return value is not a boolean", ErrPredicateReturnInvalid)
	}

	return predicateValues[0].Bool(), nil
}

func isImportLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "import")
}

func separateImports(programLines []string) ([]string, []string) {
	var importLines, nonImportLines []string

	for _, line := range programLines {
		if isImportLine(line) {
			importLines = append(importLines, line)
		} else {
			nonImportLines = append(nonImportLines, line)
		}
	}

	return importLines, nonImportLines
}

func executeTemplate(importLines, nonImportLines []string, args ...any) (bytes.Buffer, error) {
	tmpl := template.Must(template.New("main").Parse(programTemplate))
	data := templateData{Imports: strings.Join(importLines, "\n"), Program: strings.Join(nonImportLines, "\n")}

	for i, arg := range args {
		argType := reflect.TypeOf(arg)

		data.Args = append(data.Args, templateArg{
			Name: fmt.Sprintf("arg%d", i),
			Type: argType.String(),
		})
	}

	var bs bytes.Buffer

	err := tmpl.Execute(&bs, data)

	return bs, err
}

func containsReturn(lines []string) bool {
	for _, line := range lines {
		if strings.Contains(strings.TrimSpace(line), "return") {
			return true
		}
	}

	return false
}
