package match_test

import (
	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/internal"
	"github.com/grongor/go-muchtest/match"
)

const (
	actual = 3.6
)

type pkgSuite struct {
	muchtest.Suite
}

func (s *pkgSuite) matchFn(matcher match.Matcher, actual any, fn ...func(desc string)) {
	s.T().Helper()

	ok, desc := matcher.Matches(actual)

	if len(fn) == 0 {
		s.S.True(ok)
		s.S.Empty(desc)

		return
	}

	s.S.Len(1, fn)
	s.S.False(ok)
	fn[0](internal.TrimDiff(desc))
}

func (s *pkgSuite) match(matcher match.Matcher, actual any, expectedDesc ...string) {
	s.T().Helper()

	ok, desc := matcher.Matches(actual)

	if len(expectedDesc) == 0 || expectedDesc[0] == "" {
		s.S.True(ok)
		s.S.Empty(desc)

		return
	}

	s.S.Len(1, expectedDesc)
	s.S.False(ok)
	s.S.Equal(expectedDesc[0], internal.TrimDiff(desc))
}
