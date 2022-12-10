package match_test

import (
	"testing"

	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/match"
	mocks "github.com/grongor/go-muchtest/mocks/match"
)

func TestLogicSuite(t *testing.T) {
	muchtest.Run(t, new(LogicSuite))
}

type LogicSuite struct {
	pkgSuite

	Matcher1 *mocks.Matcher
	Matcher2 *mocks.Matcher
	Matcher3 *mocks.Matcher
	Matcher4 *mocks.Matcher
}

func (s *LogicSuite) BeforeTest(_, _ string) {
	s.Matcher1.EXPECT().String().Maybe().Return("T1()")
	s.Matcher2.EXPECT().String().Maybe().Return("T2()")
	s.Matcher3.EXPECT().String().Maybe().Return("T3()")
	s.Matcher4.EXPECT().String().Maybe().Return("T4()")
}

func (s *LogicSuite) TestIf_NoBranches() {
	s.match(match.If(s.Matcher1), actual,
		"Invalid If(): at least one branch (then) must be supplied. Parameters: [T1() []]")
}

func (s *LogicSuite) TestIf_TooManyBranches() {
	s.match(match.If(s.Matcher1, s.Matcher2, s.Matcher3, s.Matcher4), actual,
		"Invalid If(): at most 2 branches can be supplied (then-else). Parameters: [T1() [T2() T3() T4()]]")
}

func (s *LogicSuite) TestIf_OnlyThen() {
	s.S.Equal("If(T1(), T2())", match.If(s.Matcher1, s.Matcher2).String())

	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(match.If(s.Matcher1, s.Matcher2), actual)

	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(false, "then failed")
	s.match(match.If(s.Matcher1, s.Matcher2), actual, "then failed")

	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "if failed")
	s.match(match.If(s.Matcher1, s.Matcher2), actual, "if failed")
}

func (s *LogicSuite) TestIf() {
	s.S.Equal("If(T1(), T2(), T3())", match.If(s.Matcher1, s.Matcher2, s.Matcher3).String())

	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(match.If(s.Matcher1, s.Matcher2, s.Matcher3), actual)

	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(false, "then failed")
	s.match(match.If(s.Matcher1, s.Matcher2, s.Matcher3), actual, "then failed")

	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "if failed")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(match.If(s.Matcher1, s.Matcher2, s.Matcher3), actual)

	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "if failed")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(false, "else failed")
	s.match(match.If(s.Matcher1, s.Matcher2, s.Matcher3), actual, "else failed")
}

func (s *LogicSuite) TestNot() {
	s.S.Equal("Not(3.6)", match.Not(actual).String())
	s.S.Equal("Not(Not(Any()))", match.Not(match.Not(match.Any())).String())

	s.match(match.Not(actual), actual, "Not(3.6): matched: 3.6")
	s.match(match.Not(match.Any()), actual, "Not(Any()): matched: 3.6")
	s.match(match.Not(match.Not(match.Any())), actual)
}
