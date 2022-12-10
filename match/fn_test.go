package match_test

import (
	"regexp"
	"testing"

	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/match"
)

func TestFnSuite(t *testing.T) {
	muchtest.Run(t, new(FnSuite))
}

type FnSuite struct {
	pkgSuite
}

func (s *FnSuite) TestFn() {
	s.S.Equal("Fn()", match.Fn(func(any) bool { return true }).String())

	// without description
	s.match(match.Fn(func(actual any) bool { return true }), actual)
	s.match(match.Fn(func(actual float64) bool { return true }), actual)
	s.match(match.Fn(func(actual any) bool { return false }), actual, "Fn(): not matched value: 3.6")
	s.match(match.Fn(func(actual float64) bool { return false }), actual, "Fn(): not matched value: 3.6")

	// with description
	s.match(match.Fn(func(actual any) (bool, string) { return true, "" }), actual)
	s.match(match.Fn(func(actual float64) (bool, string) { return true, "" }), actual)
	s.match(match.Fn(func(actual any) (bool, string) { return false, "Much" }), actual, "Fn(): Much")
	s.match(match.Fn(func(actual float64) (bool, string) { return false, "Much" }), actual, "Fn(): Much")

	// invalid
	desc := "Invalid Fn(): fn must be: func(actual any|typeOf(actual)) (ok bool[, desc string]) Parameters:"
	re := regexp.MustCompile(regexp.QuoteMeta(desc) + ".*")

	for _, test := range []struct {
		name string
		fn   any
	}{
		{name: "none-none", fn: func() {}},
		{name: "any-none", fn: func(any) {}},
		{name: "any-any", fn: func(any) any { return nil }},
		{name: "int-any", fn: func(int) any { return nil }},
		{name: "int-bool", fn: func(int) bool { return true }},
		{name: "int-int", fn: func(int) int { return 1 }},
		{name: "any-int", fn: func(any) int { return 1 }},
		{name: "any-bool_any", fn: func(any) (bool, any) { return true, nil }},
		{name: "any-bool_int", fn: func(any) (bool, int) { return true, 1 }},
		{name: "int-bool_string", fn: func(int) (bool, string) { return true, "" }},
	} {
		s.Run(test.name, func() {
			s.matchFn(match.Fn(test.fn), actual, func(desc string) {
				s.S.Regexp(re, desc)
			})
		})
	}
}
