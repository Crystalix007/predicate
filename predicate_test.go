package predicate_test

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crystalix007/predicate"
)

func TestReturnTrue(t *testing.T) {
	res, err := predicate.Evaluate(getProgram("return_true"))

	require.NoError(t, err)
	assert.True(t, res)
}

func TestPanicString(t *testing.T) {
	res, err := predicate.Evaluate(getProgram("panic_string"))

	require.ErrorContains(t, err, "test")
	assert.False(t, res)
}

func TestPanicError(t *testing.T) {
	res, err := predicate.Evaluate(getProgram("panic_error"))

	require.ErrorIs(t, err, errors.ErrUnsupported)
	assert.False(t, res)
}

func TestInvalidRegexp(t *testing.T) {
	res, err := predicate.Evaluate(getProgram("invalid_regexp"))

	require.ErrorContains(t, err, "error parsing regexp")
	assert.False(t, res)
}

func TestEmptyPredicate(t *testing.T) {
	res, err := predicate.Evaluate(getProgram("empty_predicate"))

	require.ErrorIs(t, err, predicate.ErrEmptyPredicate)
	assert.False(t, res)
}

func TestInvalidPredicate(t *testing.T) {
	res, err := predicate.Evaluate(getProgram("invalid_predicate"))

	require.ErrorContains(t, err, `cannot use "string" (type stringT) as type boolT in return`)
	assert.False(t, res)
}

func TestPredicateArgs(t *testing.T) {
	res, err := predicate.Evaluate(getProgram("predicate_args"), "test")

	require.NoError(t, err)
	assert.True(t, res)
}

func getProgram(name string) string {
	filepath := path.Join("testdata", fmt.Sprintf("%s.go", name))
	bs, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	return string(bs)
}
