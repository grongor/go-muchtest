package match_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/match"
	mocks "github.com/grongor/go-muchtest/mocks/match"
)

func TestSetSuite(t *testing.T) {
	muchtest.Run(t, new(AggregateSuite))
}

type AggregateSuite struct {
	pkgSuite

	Matcher1 *mocks.Matcher
	Matcher2 *mocks.Matcher
	Matcher3 *mocks.Matcher
}

func (s *AggregateSuite) BeforeTest(_, _ string) {
	s.Matcher1.EXPECT().String().Maybe().Return("T1()")
	s.Matcher2.EXPECT().String().Maybe().Return("T2()")
	s.Matcher3.EXPECT().String().Maybe().Return("T3()")
}

func (s *AggregateSuite) TestOneOf() {
	s.S.Equal("OneOf([])", match.OneOf().String())
	s.S.Equal("OneOf([T1() T2()])", match.OneOf(s.Matcher1, s.Matcher2).String())

	// no matchers
	s.match(match.OneOf(), actual, "OneOf([]): got none: 3.6")

	// none matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.OneOf(s.Matcher1, s.Matcher2), actual, "OneOf([T1() T2()]): got none: 3.6")

	// both matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(match.OneOf(s.Matcher1, s.Matcher2), actual, "OneOf([T1() T2()]): got two: 3.6")

	// one matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.OneOf(s.Matcher1, s.Matcher2), actual)
}

func (s *AggregateSuite) TestAnyOf() {
	s.S.Equal("AnyOf([])", match.AnyOf().String())
	s.S.Equal("AnyOf([T1() T2()])", match.AnyOf(s.Matcher1, s.Matcher2).String())

	// no matchers
	s.match(match.AnyOf(), actual, "AnyOf([]): got none: 3.6")

	// none matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.AnyOf(s.Matcher1, s.Matcher2), actual, "AnyOf([T1() T2()]): got none: 3.6")

	// first matches
	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(match.AnyOf(s.Matcher1, s.Matcher2), actual)

	// second matches
	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(match.AnyOf(s.Matcher1, s.Matcher2), actual)
}

func (s *AggregateSuite) TestAtLeastOf() {
	s.S.Equal("AtLeastOf(2, [])", match.AtLeastOf(2).String())
	s.S.Equal("AtLeastOf(2, [T1() T2() T3()])", match.AtLeastOf(2, s.Matcher1, s.Matcher2, s.Matcher3).String())

	// invalid
	s.match(match.AtLeastOf(0), actual, "Invalid AtLeastOf(): n can't be less than 1. Parameters: [0 []]")

	// no matchers
	s.match(match.AtLeastOf(2), actual, "AtLeastOf(2, []): got none: 3.6")

	// none matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.AtLeastOf(2, s.Matcher1, s.Matcher2, s.Matcher3), actual,
		"AtLeastOf(2, [T1() T2() T3()]): got none: 3.6")

	// not enough matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.AtLeastOf(2, s.Matcher1, s.Matcher2, s.Matcher3), actual,
		"AtLeastOf(2, [T1() T2() T3()]): got one: 3.6")

	// enough matches
	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(match.AtLeastOf(2, s.Matcher1, s.Matcher2, s.Matcher3), actual)
}

func (s *AggregateSuite) TestAtMostOf() {
	s.S.Equal("AtMostOf(2, [])", match.AtMostOf(2).String())
	s.S.Equal("AtMostOf(2, [T1() T2() T3()])", match.AtMostOf(2, s.Matcher1, s.Matcher2, s.Matcher3).String())

	// invalid
	s.match(match.AtMostOf(-1), actual, "Invalid AtMostOf(): n can't be less than 0. Parameters: [-1 []]")

	// no matchers
	s.match(match.AtMostOf(2), actual)

	// none matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.AtMostOf(2, s.Matcher1, s.Matcher2, s.Matcher3), actual)

	// one matches
	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.AtMostOf(2, s.Matcher1, s.Matcher2, s.Matcher3), actual)

	// two matches
	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.AtMostOf(2, s.Matcher1, s.Matcher2, s.Matcher3), actual)

	// three matches
	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(match.AtMostOf(2, s.Matcher1, s.Matcher2, s.Matcher3), actual,
		"AtMostOf(2, [T1() T2() T3()]): got three: 3.6")

	// large number formatting
	matchers := make([]any, 13)
	for i := 0; i < 13; i++ {
		matchers[i] = s.Matcher1
	}

	s.Matcher1.EXPECT().Matches(actual).Times(13).Return(true, "")
	s.matchFn(match.AtMostOf(0, matchers...), actual, func(desc string) {
		msg := "AtMostOf(0, [T1() T1() T1() T1() T1() T1() T1() T1() T1() T1() T1() T1() T1()]): got 13: 3.6"
		s.S.Equal(msg, desc)
		s.S.Equal(fmt.Sprintf("AtMostOf(0, [%s]): got 13: 3.6", strings.Repeat(" T1()", 13)[1:]), desc)
	})
}

func (s *AggregateSuite) TestBetweenOf() {
	s.S.Equal("BetweenOf(1, 2, [])", match.BetweenOf(1, 2).String())
	s.S.Equal("BetweenOf(1, 2, [T1() T2() T3()])", match.BetweenOf(1, 2, s.Matcher1, s.Matcher2, s.Matcher3).String())

	// invalid
	s.match(match.BetweenOf(-1, 2), actual, "Invalid BetweenOf(): min can't be less than 1. Parameters: [-1 2 []]")

	// invalid
	s.match(match.BetweenOf(3, 2), actual, "Invalid BetweenOf(): min can't be more than max. Parameters: [3 2 []]")

	// no matchers
	s.match(match.BetweenOf(1, 2), actual, "BetweenOf(1, 2, []): got none: 3.6")

	// none matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.BetweenOf(1, 2, s.Matcher1, s.Matcher2, s.Matcher3), actual,
		"BetweenOf(1, 2, [T1() T2() T3()]): got none: 3.6")

	// one matches
	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.BetweenOf(1, 2, s.Matcher1, s.Matcher2, s.Matcher3), actual)

	// two matches
	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(false, "")
	s.match(match.BetweenOf(1, 2, s.Matcher1, s.Matcher2, s.Matcher3), actual)

	// three matches
	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher3.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(match.BetweenOf(1, 2, s.Matcher1, s.Matcher2, s.Matcher3), actual,
		"BetweenOf(1, 2, [T1() T2() T3()]): got three: 3.6")
}

func (s *AggregateSuite) TestAll() {
	s.S.Equal("All([])", match.All().String())
	s.S.Equal("All([T1() T2()])", match.All(s.Matcher1, s.Matcher2).String())

	s.Matcher1.EXPECT().String().Maybe().Return("T1()")
	s.Matcher2.EXPECT().String().Maybe().Return("T2()")

	// no matchers
	s.match(match.All(), actual)

	// one, not matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "nok")
	s.match(match.All(s.Matcher1), actual, "All([T1()]): got none: nok")
	s.match(match.All("Much"), actual, `All([Much]): got none: Equal("Much"): not equal: 3.6`)

	// one, matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(match.All(s.Matcher1), actual)
	s.match(match.All(actual), actual)

	// two, one matched
	matcher := match.All(s.Matcher1, s.Matcher2)

	s.Matcher1.EXPECT().Matches(actual).Once().Return(false, "nok")
	s.match(matcher, actual, "All([T1() T2()]): at least one (index 0) failed: nok")

	// two, both matched
	s.Matcher1.EXPECT().Matches(actual).Once().Return(true, "")
	s.Matcher2.EXPECT().Matches(actual).Once().Return(true, "")
	s.match(matcher, actual)
}
