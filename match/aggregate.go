package match

import (
	"fmt"
	"math"
	"strconv"
)

func AnyOf(expected ...any) StatefulMatcher {
	return newNOfMatcher("AnyOf", expected, 1, math.MaxInt)
}

func OneOf(expected ...any) StatefulMatcher {
	return newNOfMatcher("OneOf", expected, 1, 1)
}

func AtLeastOf(n int, expected ...any) StatefulMatcher {
	if n < 1 {
		return matcherErr("Invalid AtLeastOf(): n can't be less than 1.", n, expected)
	}

	return newNOfMatcher("AtLeastOf", expected, n, math.MaxInt)
}

func AtMostOf(n int, expected ...any) StatefulMatcher {
	if n < 0 {
		return matcherErr("Invalid AtMostOf(): n can't be less than 0.", n, expected)
	}

	return newNOfMatcher("AtMostOf", expected, 0, n)
}

func BetweenOf(min int, max int, expected ...any) StatefulMatcher {
	if min < 1 {
		return matcherErr("Invalid BetweenOf(): min can't be less than 1.", min, max, expected)
	}

	if min > max {
		return matcherErr("Invalid BetweenOf(): min can't be more than max.", min, max, expected)
	}

	return newNOfMatcher("BetweenOf", expected, min, max)
}

func All(expected ...any) Matcher {
	matcher := newNOfMatcher("All", expected, len(expected), len(expected))
	matcher.all = true

	return matcher
}

type aggregateMatcher struct {
	name                     string
	expected                 []Matcher
	min, max                 int
	all, unlimited, stateful bool
}

func (m *aggregateMatcher) Matches(actual any) (ok bool, desc string) {
	count := 0

	for i := 0; i < len(m.expected); i++ {
		if ok, desc = m.expected[i].Matches(actual); !ok {
			if m.all {
				failed := ""
				if len(m.expected)-count > 1 {
					failed = fmt.Sprintf("at least one (index %d) failed", i)
				} else {
					failed = "got " + m.itoa(count)
				}

				//m.itoa(len(m.expected))
				return false, fmt.Sprintf("%s: %s: %s", m.String(), failed, desc)
			}

			continue
		}

		if m.stateful {
			m.expected = append(m.expected[:i], m.expected[i+1:]...)
			i--
		}

		if count++; m.unlimited && count == m.min {
			return true, ""
		}
	}

	if count >= m.min && count <= m.max {
		return true, ""
	}

	return false, fmt.Sprintf("%s: got %s: %s", m.String(), m.itoa(count), formatValue(actual))
}

func (m *aggregateMatcher) String() string {
	return m.name
}

func (m *aggregateMatcher) Stateful() Matcher {
	m.stateful = true

	return m
}

func (m *aggregateMatcher) itoa(n int) string {
	if n > 12 {
		return strconv.Itoa(n)
	}

	numbers := []string{"none", "one", "two", "three", "four", "five",
		"six", "seven", "eight", "nine", "ten", "eleven", "twelve"}

	return numbers[n]
}

func newNOfMatcher(name string, expected []any, min, max int) *aggregateMatcher {
	matcher := &aggregateMatcher{expected: ToMatchers(expected), min: min, max: max}

	var maxStr string
	if max == math.MaxInt {
		matcher.unlimited = true
		maxStr = "inf"
	} else {
		maxStr = strconv.Itoa(max)
	}

	switch name {
	case "AnyOf", "OneOf", "All":
		matcher.name = fmt.Sprintf("%s(%v)", name, expected)
	case "AtLeastOf":
		matcher.name = fmt.Sprintf("%s(%d, %v)", name, min, expected)
	case "AtMostOf":
		matcher.name = fmt.Sprintf("%s(%s, %v)", name, maxStr, expected)
	case "BetweenOf":
		matcher.name = fmt.Sprintf("%s(%d, %s, %v)", name, min, maxStr, expected)
	}

	return matcher
}
