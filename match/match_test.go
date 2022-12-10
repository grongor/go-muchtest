package match_test

import (
	"testing"

	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/match"
)

func TestMatchSuite(t *testing.T) {
	muchtest.Run(t, new(MatchSuite))
}

type MatchSuite struct {
	pkgSuite
}

func (s *MatchSuite) TestAnyAndNone() {
	s.S.Equal("Any()", match.Any().String())
	s.S.Equal("None()", match.None().String())

	for _, test := range []struct {
		name   string
		actual any
	}{
		{name: "nil", actual: nil},
		{name: "int", actual: 0},
		{name: "string", actual: "Much"},
		{name: "int64", actual: int64(0)},
		{name: "struct{}", actual: struct{}{}},
	} {
		s.Run(test.name, func() {
			s.match(match.Any(), test.actual)
			s.match(match.None(), test.actual, "None() never matches")
		})
	}
}
