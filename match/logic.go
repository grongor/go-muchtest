package match

import (
	"fmt"
)

func If(condition Matcher, branches ...any) Matcher {
	switch len(branches) {
	case 0:
		return matcherErr("Invalid If(): at least one branch (then) must be supplied.", condition, branches)
	case 1:
		return condMatcher{condition: condition, success: ToMatcher(branches[0])}
	case 2:
		return condMatcher{condition: condition, success: ToMatcher(branches[0]), failure: ToMatcher(branches[1])}
	default:
		return matcherErr("Invalid If(): at most 2 branches can be supplied (then-else).", condition, branches)
	}
}

func Not(expected any) Matcher {
	return notMatcher{not: expected}
}

type condMatcher struct {
	condition, success, failure Matcher
}

func (m condMatcher) Matches(actual any) (ok bool, desc string) {
	if ok, desc = m.condition.Matches(actual); ok {
		return m.success.Matches(actual)
	} else if m.failure == nil {
		return ok, desc
	}

	return m.failure.Matches(actual)
}

func (m condMatcher) String() string {
	if m.failure == nil {
		return fmt.Sprintf("If(%v, %v)", formatValue(m.condition), formatValue(m.success))
	}

	return fmt.Sprintf("If(%v, %v, %v)", formatValue(m.condition), formatValue(m.success), formatValue(m.failure))
}

type notMatcher struct {
	not any
}

func (m notMatcher) Matches(actual any) (ok bool, desc string) {
	if ok, _ = ToMatcher(m.not).Matches(actual); ok {
		return false, fmt.Sprintf("%s: matched: %s", m.String(), formatValue(actual))
	}

	return true, ""
}

func (m notMatcher) String() string {
	return fmt.Sprintf("Not(%s)", formatValue(m.not))
}
