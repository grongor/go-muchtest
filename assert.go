package muchtest

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grongor/go-muchtest/match"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Assertions struct {
	t TestingT
}

func (a *Assertions) Fail(message string, messageAndArgs ...any) {
	a.t.Helper()

	require.Fail(a.t, message, messageAndArgs...)
}

func (a *Assertions) Match(matcher match.Matcher, actual any, messageAndArgs ...any) {
	a.t.Helper()

	if ok, desc := matcher.Matches(actual); !ok {
		require.Fail(a.t, desc, messageAndArgs...)
	}
}

func (a *Assertions) Equal(expected, actual any, messageAndArgs ...any) {
	a.t.Helper()

	a.EqualOpt(nil, expected, actual, messageAndArgs...)
}

func (a *Assertions) NotEqual(expected, actual any, messageAndArgs ...any) {
	a.t.Helper()

	a.NotEqualOpt(nil, expected, actual, messageAndArgs...)
}

func (a *Assertions) EqualOpt(options []cmp.Option, expected, actual any, messageAndArgs ...any) {
	a.t.Helper()

	if ok, desc := match.Equal(expected, options...).Matches(actual); !ok {
		require.Fail(a.t, desc, messageAndArgs)
	}
}

func (a *Assertions) NotEqualOpt(options []cmp.Option, expected, actual any, messageAndArgs ...any) {
	a.t.Helper()

	if ok, desc := match.Equal(expected, options...).Matches(actual); ok {
		require.Fail(a.t, desc, messageAndArgs)
	}
}

func (a *Assertions) Nil(actual any, messageAndArgs ...any) {
	a.t.Helper()

	if actual != nil {
		require.Fail(a.t, fmt.Sprintf("Expected nil value, got: %#v", actual), messageAndArgs)
	}
}

func (a *Assertions) NotNil(actual any, messageAndArgs ...any) {
	a.t.Helper()

	if actual == nil {
		require.Fail(a.t, "Expected not nil value", messageAndArgs)
	}
}

func (a *Assertions) Empty(actual any, messageAndArgs ...any) {
	require.Empty(a.t, actual, messageAndArgs)
}

func (a *Assertions) True(actual bool, messageAndArgs ...any) {
	require.True(a.t, actual, messageAndArgs...)
}

func (a *Assertions) False(actual bool, messageAndArgs ...any) {
	require.False(a.t, actual, messageAndArgs...)
}

func (a *Assertions) Len(n int, actual any, messageAndArgs ...any) {
	if ok, desc := match.Len(n).Matches(actual); !ok {
		require.Fail(a.t, desc, messageAndArgs)
	}
}

func (a *Assertions) Error(expected any, actual error, messageAndArgs ...any) {
	switch e := expected.(type) {
	case error:
		if e != actual && e.Error() != actual.Error() && !errors.Is(e, actual) {
			require.Fail(a.t, fmt.Sprintf("Expected different error: %#v", actual), messageAndArgs...)
		}
	case string:
		if strings.Contains(actual.Error(), e) {
			require.Fail(a.t, fmt.Sprintf(`Expected different error message "%s": %s`, e, actual), messageAndArgs...)
		}
	case match.Matcher:
		if ok, desc := e.Matches(actual); !ok {
			require.Fail(a.t, fmt.Sprintf("Error doesn't match expectations: %s: %#v", desc, actual))
		}
	default:
		if expected == nil {
			if actual == nil {
				require.Fail(a.t, "Expected any error, got nil")
			}

			return
		}

		a.Equal(expected, actual, messageAndArgs...)
	}
}

func (a *Assertions) NoError(actual error, messageAndArgs ...any) {
	require.NoError(a.t, actual, messageAndArgs...)
}

func (a *Assertions) Regexp(regexp, actual any, messageAndArgs ...any) {
	if ok, desc := match.Regexp(regexp).Matches(actual); !ok {
		require.Fail(a.t, desc, messageAndArgs...)
	}
}

///////////////////////////////////////////////////////////////////////////////////////

type TestifyInterface interface {
	// Contains asserts that the specified string, list(array, slice...) or map contains the
	// specified substring or element.
	//
	//	assert.Contains(t, "Hello World", "World")
	//	assert.Contains(t, ["Hello", "World"], "World")
	//	assert.Contains(t, {"Hello": "World"}, "Hello")
	Contains(t TestingT, s interface{}, contains interface{}, msgAndArgs ...interface{})

	// DirExists checks whether a directory exists in the given path. It also fails
	// if the path is a file rather a directory or there is an error checking whether it exists.
	DirExists(t TestingT, path string, msgAndArgs ...interface{})

	// ElementsMatch asserts that the specified listA(array, slice...) is equal to specified
	// listB(array, slice...) ignoring the order of the elements. If there are duplicate elements,
	// the number of appearances of each of them in both lists should match.
	//
	// assert.ElementsMatch(t, [1, 3, 2, 3], [1, 3, 3, 2])
	ElementsMatch(t TestingT, listA interface{}, listB interface{}, msgAndArgs ...interface{})

	// Empty asserts that the specified object is empty.  I.e. nil, "", false, 0 or either
	// a slice or a channel with len == 0.
	//
	//	assert.Empty(t, obj)
	Empty(t TestingT, object interface{}, msgAndArgs ...interface{})

	// Equal asserts that two objects are equal.
	//
	//	assert.Equal(t, 123, 123)
	//
	// Pointer variable equality is determined based on the equality of the
	// referenced values (as opposed to the memory addresses). Function equality
	// cannot be determined and will always fail.
	Equal(t TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{})

	// EqualError asserts that a function returned an error (i.e. not `nil`)
	// and that it is equal to the provided error.
	//
	//	actualObj, err := SomeFunction()
	//	assert.EqualError(t, err,  expectedErrorString)
	EqualError(t TestingT, theError error, errString string, msgAndArgs ...interface{})
	// EqualValues asserts that two objects are equal or convertable to the same types
	// and equal.
	//
	//	assert.EqualValues(t, uint32(123), int32(123))
	EqualValues(t TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{})

	// Error asserts that a function returned an error (i.e. not `nil`).
	//
	//	  actualObj, err := SomeFunction()
	//	  if assert.Error(t, err) {
	//		   assert.Equal(t, expectedError, err)
	//	  }
	Error(t TestingT, err error, msgAndArgs ...interface{})

	// ErrorAs asserts that at least one of the errors in err's chain matches target, and if so, sets target to that error value.
	// This is a wrapper for errors.As.
	ErrorAs(t TestingT, err error, target interface{}, msgAndArgs ...interface{})

	// ErrorContains asserts that a function returned an error (i.e. not `nil`)
	// and that the error contains the specified substring.
	//
	//	actualObj, err := SomeFunction()
	//	assert.ErrorContains(t, err,  expectedErrorSubString)
	ErrorContains(t TestingT, theError error, contains string, msgAndArgs ...interface{})

	// ErrorIs asserts that at least one of the errors in err's chain matches target.
	// This is a wrapper for errors.Is.
	ErrorIs(t TestingT, err error, target error, msgAndArgs ...interface{})

	// Eventually asserts that given condition will be met in waitFor time,
	// periodically checking target function each tick.
	//
	//	assert.Eventually(t, func() bool { return true; }, time.Second, 10*time.Millisecond)
	Eventually(t TestingT, condition func() bool, waitFor time.Duration, tick time.Duration, msgAndArgs ...interface{})

	// Exactly asserts that two objects are equal in value and type.
	//
	//	assert.Exactly(t, int32(123), int64(123))
	Exactly(t TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{})

	// Fail reports a failure through
	Fail(t TestingT, failureMessage string, msgAndArgs ...interface{})

	// FailNow fails test
	FailNow(t TestingT, failureMessage string, msgAndArgs ...interface{})

	// False asserts that the specified value is false.
	//
	//	assert.False(t, myBool)
	False(t TestingT, value bool, msgAndArgs ...interface{})

	// Greater asserts that the first element is greater than the second
	//
	//	assert.Greater(t, 2, 1)
	//	assert.Greater(t, float64(2), float64(1))
	//	assert.Greater(t, "b", "a")
	Greater(t TestingT, e1 interface{}, e2 interface{}, msgAndArgs ...interface{})

	// GreaterOrEqual asserts that the first element is greater than or equal to the second
	//
	//	assert.GreaterOrEqual(t, 2, 1)
	//	assert.GreaterOrEqual(t, 2, 2)
	//	assert.GreaterOrEqual(t, "b", "a")
	//	assert.GreaterOrEqual(t, "b", "b")
	GreaterOrEqual(t TestingT, e1 interface{}, e2 interface{}, msgAndArgs ...interface{})

	// Implements asserts that an object is implemented by the specified interface.
	//
	//	assert.Implements(t, (*MyInterface)(nil), new(MyObject))
	Implements(t TestingT, interfaceObject interface{}, object interface{}, msgAndArgs ...interface{})

	// InDelta asserts that the two numerals are within delta of each other.
	//
	//	assert.InDelta(t, math.Pi, 22/7.0, 0.01)
	InDelta(t TestingT, expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{})

	// InDeltaMapValues is the same as InDelta, but it compares all values between two maps. Both maps must have exactly the same keys.
	InDeltaMapValues(t TestingT, expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{})

	// InDeltaSlice is the same as InDelta, except it compares two slices.
	InDeltaSlice(t TestingT, expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{})

	// InEpsilon asserts that expected and actual have a relative error less than epsilon
	InEpsilon(t TestingT, expected interface{}, actual interface{}, epsilon float64, msgAndArgs ...interface{})

	// InEpsilonSlice is the same as InEpsilon, except it compares each value from two slices.
	InEpsilonSlice(t TestingT, expected interface{}, actual interface{}, epsilon float64, msgAndArgs ...interface{})

	// IsDecreasing asserts that the collection is decreasing
	//
	//	assert.IsDecreasing(t, []int{2, 1, 0})
	//	assert.IsDecreasing(t, []float{2, 1})
	//	assert.IsDecreasing(t, []string{"b", "a"})
	IsDecreasing(t TestingT, object interface{}, msgAndArgs ...interface{})

	// IsIncreasing asserts that the collection is increasing
	//
	//	assert.IsIncreasing(t, []int{1, 2, 3})
	//	assert.IsIncreasing(t, []float{1, 2})
	//	assert.IsIncreasing(t, []string{"a", "b"})
	IsIncreasing(t TestingT, object interface{}, msgAndArgs ...interface{})

	// IsNonDecreasing asserts that the collection is not decreasing
	//
	//	assert.IsNonDecreasing(t, []int{1, 1, 2})
	//	assert.IsNonDecreasing(t, []float{1, 2})
	//	assert.IsNonDecreasing(t, []string{"a", "b"})
	IsNonDecreasing(t TestingT, object interface{}, msgAndArgs ...interface{})

	// IsNonIncreasing asserts that the collection is not increasing
	//
	//	assert.IsNonIncreasing(t, []int{2, 1, 1})
	//	assert.IsNonIncreasing(t, []float{2, 1})
	//	assert.IsNonIncreasing(t, []string{"b", "a"})
	IsNonIncreasing(t TestingT, object interface{}, msgAndArgs ...interface{})

	// IsType asserts that the specified objects are of the same type.
	IsType(t TestingT, expectedType interface{}, object interface{}, msgAndArgs ...interface{})

	// Len asserts that the specified object has specific length.
	// Len also fails if the object has a type that len() not accept.
	//
	//	assert.Len(t, mySlice, 3)
	Len(t TestingT, object interface{}, length int, msgAndArgs ...interface{})

	// Less asserts that the first element is less than the second
	//
	//	assert.Less(t, 1, 2)
	//	assert.Less(t, float64(1), float64(2))
	//	assert.Less(t, "a", "b")
	Less(t TestingT, e1 interface{}, e2 interface{}, msgAndArgs ...interface{})

	// LessOrEqual asserts that the first element is less than or equal to the second
	//
	//	assert.LessOrEqual(t, 1, 2)
	//	assert.LessOrEqual(t, 2, 2)
	//	assert.LessOrEqual(t, "a", "b")
	//	assert.LessOrEqual(t, "b", "b")
	LessOrEqual(t TestingT, e1 interface{}, e2 interface{}, msgAndArgs ...interface{})

	// Negative asserts that the specified element is negative
	//
	//	assert.Negative(t, -1)
	//	assert.Negative(t, -1.23)
	Negative(t TestingT, e interface{}, msgAndArgs ...interface{})

	// Never asserts that the given condition doesn't satisfy in waitFor time,
	// periodically checking the target function each tick.
	//
	//	assert.Never(t, func() bool { return false; }, time.Second, 10*time.Millisecond)
	Never(t TestingT, condition func() bool, waitFor time.Duration, tick time.Duration, msgAndArgs ...interface{})

	// Nil asserts that the specified object is nil.
	//
	//	assert.Nil(t, err)
	Nil(t TestingT, object interface{}, msgAndArgs ...interface{})

	// NoError asserts that a function returned no error (i.e. `nil`).
	//
	//	  actualObj, err := SomeFunction()
	//	  if assert.NoError(t, err) {
	//		   assert.Equal(t, expectedObj, actualObj)
	//	  }
	NoError(t TestingT, err error, msgAndArgs ...interface{})

	// NotContains asserts that the specified string, list(array, slice...) or map does NOT contain the
	// specified substring or element.
	//
	//	assert.NotContains(t, "Hello World", "Earth")
	//	assert.NotContains(t, ["Hello", "World"], "Earth")
	//	assert.NotContains(t, {"Hello": "World"}, "Earth")
	NotContains(t TestingT, s interface{}, contains interface{}, msgAndArgs ...interface{})

	// NotEmpty asserts that the specified object is NOT empty.  I.e. not nil, "", false, 0 or either
	// a slice or a channel with len == 0.
	//
	//	if assert.NotEmpty(t, obj) {
	//	  assert.Equal(t, "two", obj[1])
	//	}
	NotEmpty(t TestingT, object interface{}, msgAndArgs ...interface{})

	// NotEqual asserts that the specified values are NOT equal.
	//
	//	assert.NotEqual(t, obj1, obj2)
	//
	// Pointer variable equality is determined based on the equality of the
	// referenced values (as opposed to the memory addresses).
	NotEqual(t TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{})

	// NotEqualValues asserts that two objects are not equal even when converted to the same type
	//
	//	assert.NotEqualValues(t, obj1, obj2)
	NotEqualValues(t TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{})

	// NotErrorIs asserts that at none of the errors in err's chain matches target.
	// This is a wrapper for errors.Is.
	NotErrorIs(t TestingT, err error, target error, msgAndArgs ...interface{})

	// NotNil asserts that the specified object is not nil.
	//
	//	assert.NotNil(t, err)
	NotNil(t TestingT, object interface{}, msgAndArgs ...interface{})

	// NotPanics asserts that the code inside the specified PanicTestFunc does NOT panic.
	//
	//	assert.NotPanics(t, func(){ RemainCalm() })
	NotPanics(t TestingT, f assert.PanicTestFunc, msgAndArgs ...interface{})

	// NotRegexp asserts that a specified regexp does not match a string.
	//
	//	assert.NotRegexp(t, regexp.MustCompile("starts"), "it's starting")
	//	assert.NotRegexp(t, "^start", "it's not starting")
	NotRegexp(t TestingT, rx interface{}, str interface{}, msgAndArgs ...interface{})

	// NotSame asserts that two pointers do not reference the same object.
	//
	//	assert.NotSame(t, ptr1, ptr2)
	//
	// Both arguments must be pointer variables. Pointer variable sameness is
	// determined based on the equality of both type and value.
	NotSame(t TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{})

	// NotSubset asserts that the specified list(array, slice...) contains not all
	// elements given in the specified subset(array, slice...).
	//
	//	assert.NotSubset(t, [1, 3, 4], [1, 2], "But [1, 3, 4] does not contain [1, 2]")
	NotSubset(t TestingT, list interface{}, subset interface{}, msgAndArgs ...interface{})

	// NotZero asserts that i is not the zero value for its type.
	NotZero(t TestingT, i interface{}, msgAndArgs ...interface{})

	// Panics asserts that the code inside the specified PanicTestFunc panics.
	//
	//	assert.Panics(t, func(){ GoCrazy() })
	Panics(t TestingT, f assert.PanicTestFunc, msgAndArgs ...interface{})

	// PanicsWithError asserts that the code inside the specified PanicTestFunc
	// panics, and that the recovered panic value is an error that satisfies the
	// EqualError comparison.
	//
	//	assert.PanicsWithError(t, "crazy error", func(){ GoCrazy() })
	PanicsWithError(t TestingT, errString string, f assert.PanicTestFunc, msgAndArgs ...interface{})

	// PanicsWithValue asserts that the code inside the specified PanicTestFunc panics, and that
	// the recovered panic value equals the expected panic value.
	//
	//	assert.PanicsWithValue(t, "crazy error", func(){ GoCrazy() })
	PanicsWithValue(t TestingT, expected interface{}, f assert.PanicTestFunc, msgAndArgs ...interface{})

	// Positive asserts that the specified element is positive
	//
	//	assert.Positive(t, 1)
	//	assert.Positive(t, 1.23)
	Positive(t TestingT, e interface{}, msgAndArgs ...interface{})

	// Regexp asserts that a specified regexp matches a string.
	//
	//	assert.Regexp(t, regexp.MustCompile("start"), "it's starting")
	//	assert.Regexp(t, "start...$", "it's not starting")
	Regexp(t TestingT, rx interface{}, str interface{}, msgAndArgs ...interface{})

	// Same asserts that two pointers reference the same object.
	//
	//	assert.Same(t, ptr1, ptr2)
	//
	// Both arguments must be pointer variables. Pointer variable sameness is
	// determined based on the equality of both type and value.
	Same(t TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{})

	// Subset asserts that the specified list(array, slice...) contains all
	// elements given in the specified subset(array, slice...).
	//
	//	assert.Subset(t, [1, 2, 3], [1, 2], "But [1, 2, 3] does contain [1, 2]")
	Subset(t TestingT, list interface{}, subset interface{}, msgAndArgs ...interface{})

	// True asserts that the specified value is true.
	//
	//	assert.True(t, myBool)
	True(t TestingT, value bool, msgAndArgs ...interface{})

	// Zero asserts that i is the zero value for its type.
	Zero(t TestingT, i interface{}, msgAndArgs ...interface{})
}
