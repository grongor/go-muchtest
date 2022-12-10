package match_test

import (
	"testing"

	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/match"
)

func TestContainsSuite(t *testing.T) {
	muchtest.Run(t, new(ContainsSuite))
}

type ContainsSuite struct {
	pkgSuite
}

func (s *ContainsSuite) TestContains_String() {
	actual := "Do Much Test"

	s.match(match.Contains("Much"), actual)
	s.match(match.Contains([]byte("Much")), actual)

	s.match(match.Contains("Such"), actual, `Contains("Such"): not contained in: "Do Much Test"`)
	s.match(match.Contains([]byte("Such")), actual, `Contains("Such"): not contained in: "Do Much Test"`)
	s.match(match.Contains(match.Any()), actual,
		"Contains(Any()): value can't be a Matcher when typeOf(actual) == string")
	s.match(match.Contains(123), actual,
		"Contains(123): invalid value; possible types: string, fmt.Stringer, []byte, time.Time, *regexp.Regexp")
	s.match(match.ContainsKey("whatever"), actual,
		`ContainsKey("whatever"): can't be use when typeOf(actual) == string`)
	s.match(match.ContainsKeyValue("whatever", "Much"), actual,
		`ContainsKeyValue("whatever", "Much"): can't be use when typeOf(actual) == string`)
}

func (s *ContainsSuite) TestContains_Slice() {
	actual := []int{2, 0, 8}

	// only by value
	s.match(match.Contains(2), actual)
	s.match(match.Contains(match.Any()), actual)
	s.match(match.Contains(3), actual, "Contains(3): no such value in: []int{2, 0, 8}")
	s.match(match.Contains("Much"), actual, `Contains("Much"): no such value in: []int{2, 0, 8}`)

	// only by key
	s.match(match.ContainsKey(2), actual)
	s.match(match.ContainsKey(3), actual, "ContainsKey(3): key out of range: len(actual) == 3")
	s.match(match.ContainsKey("Much"), actual, `ContainsKey("Much"): key must be int when typeOf(actual) == []any`)
	s.match(match.ContainsKey(match.Any()), actual, "ContainsKey(Any()): key must be int when typeOf(actual) == []any")
	s.match(match.ContainsKey(1), actual, "ContainsKey(1): value is zero: 0")

	// by key and value
	s.match(match.ContainsKeyValue(1, 0), actual)
	s.match(match.ContainsKeyValue(1, match.Any()), actual)
	s.match(match.ContainsKeyValue("Much", match.Any()), actual,
		`ContainsKeyValue("Much", Any()): key must be int when typeOf(actual) == []any`)
	s.match(match.ContainsKeyValue(3, match.Any()), actual,
		"ContainsKeyValue(3, Any()): key out of range: len(actual) == 3")
	s.match(match.ContainsKeyValue(1, 1), actual, "ContainsKeyValue(1, 1): value not matched: 0")
	s.match(match.ContainsKeyValue(1, "Much"), actual, `ContainsKeyValue(1, "Much"): value not matched: 0`)
	s.match(match.ContainsKeyValue(1, match.None()), actual, "ContainsKeyValue(1, None()): value not matched: 0")
}

func (s *ContainsSuite) TestContains_Map() {
	actual := map[int]string{0: "So", 2: "Much", 4: "Test"}

	// only by value
	s.match(match.Contains("Much"), actual)
	s.match(match.Contains(match.Any()), actual)
	s.match(match.Contains("Not"), actual,
		`Contains("Not"): no such value in: map[int]string{0:"So", 2:"Much", 4:"Test"}`)
	s.match(match.Contains(2), actual, `Contains(2): no such value in: map[int]string{0:"So", 2:"Much", 4:"Test"}`)

	// only by key
	s.match(match.ContainsKey(2), actual)
	s.match(match.ContainsKey(match.Any()), actual)
	s.match(match.ContainsKey(3), actual,
		`ContainsKey(3): no such key in: map[int]string{0:"So", 2:"Much", 4:"Test"}`)
	s.match(match.ContainsKey("Much"), actual,
		`ContainsKey("Much"): no such key in: map[int]string{0:"So", 2:"Much", 4:"Test"}`)

	// by key and value
	s.match(match.ContainsKeyValue(2, "Much"), actual)
	s.match(match.ContainsKeyValue(2, match.Any()), actual)
	s.match(match.ContainsKeyValue(match.Any(), match.AnyOf("So", "Much", "Test")), actual)
	s.match(match.ContainsKeyValue(1, match.Any()), actual,
		`ContainsKeyValue(1, Any()): no such key in: map[int]string{0:"So", 2:"Much", 4:"Test"}`)
	s.match(match.ContainsKeyValue("Much", match.Any()), actual,
		`ContainsKeyValue("Much", Any()): no such key in: map[int]string{0:"So", 2:"Much", 4:"Test"}`)
	s.match(match.ContainsKeyValue(2, "Such"), actual, `ContainsKeyValue(2, "Such"): value not matched: "Much"`)
	s.match(match.ContainsKeyValue(2, 1), actual, `ContainsKeyValue(2, 1): value not matched: "Much"`)
	s.match(match.ContainsKeyValue(2, match.None()), actual, `ContainsKeyValue(2, None()): value not matched: "Much"`)
}

func (s *ContainsSuite) TestContains_Struct() {
	const structStr = `match_test.outer{Much:"Very", Test:match_test.inner{Deep:3.6}}`
	type inner struct {
		lorem int
		Deep  float64
	}
	type outer struct {
		unexported string
		Much       string
		Test       inner
	}
	actual := outer{
		unexported: "hello",
		Much:       "Very",
		Test: inner{
			lorem: 123,
			Deep:  actual,
		},
	}

	// only by value
	s.match(match.Contains("Very"), actual)
	s.match(match.Contains(match.Any()), actual)
	s.match(match.Contains("hello"), actual, `Contains("hello"): no such value in: `+structStr)
	s.match(match.Contains(2), actual, `Contains(2): no such value in: `+structStr)

	// only by key
	s.match(match.ContainsKey("Much"), actual)
	s.match(match.ContainsKey(match.Any()), actual)
	s.match(match.ContainsKey(3), actual, `ContainsKey(3): no such field in: `+structStr)
	s.match(match.ContainsKey("Such"), actual, `ContainsKey("Such"): no such field in: `+structStr)

	// by key and value
	s.match(match.ContainsKeyValue("Much", "Very"), actual)
	s.match(match.ContainsKeyValue("Test", match.Any()), actual)
	s.match(match.ContainsKeyValue(match.Any(), match.AnyOf("hello", "Very")), actual)
	s.match(match.ContainsKeyValue(1, match.Any()), actual,
		`ContainsKeyValue(1, Any()): no such field in: `+structStr)
	s.match(match.ContainsKeyValue("Such", match.Any()), actual,
		`ContainsKeyValue("Such", Any()): no such field in: `+structStr)
	s.match(match.ContainsKeyValue("Much", "Such"), actual,
		`ContainsKeyValue("Much", "Such"): value not matched: "Very"`)
	s.match(match.ContainsKeyValue("Much", 1), actual, `ContainsKeyValue("Much", 1): value not matched: "Very"`)
	s.match(match.ContainsKeyValue("Much", match.None()), actual,
		`ContainsKeyValue("Much", None()): value not matched: "Very"`)
}

func (s *ContainsSuite) TestContains() {
	s.S.Equal(`Contains("Much")`, match.Contains("Much").String())
	s.S.Equal(`ContainsKey("Much")`, match.ContainsKey("Much").String())
	s.S.Equal(`ContainsKeyValue("Much", 123)`, match.ContainsKeyValue("Much", 123).String())

	s.match(match.Contains(""), actual, `Contains(""): unsupported container kind: float64`)
	s.match(match.ContainsKey(""), actual, `ContainsKey(""): unsupported container kind: float64`)
	s.match(match.ContainsKeyValue("", ""), actual, `ContainsKeyValue("", ""): unsupported container kind: float64`)
}
