package match_test

import (
	"strings"
	"testing"

	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/match"
)

func TestLenSuite(t *testing.T) {
	muchtest.Run(t, new(LenSuite))
}

type LenSuite struct {
	pkgSuite
}

func (s *LenSuite) TestLen() {
	s.match(match.Len(0), []string(nil))
	s.match(match.Len(0), []string{})
	s.match(match.Len(1), []string{"Much"})
	s.match(match.Len(2), []string{"Much"}, `Len(2): got 1: []string{"Much"}`)
	s.matchFn(match.Len(2), func() {}, func(desc string) {
		s.S.True(strings.HasPrefix(desc, "Len(2): len() panicked for value: (func())"))
	})
}
