package match

import (
	"fmt"
	"regexp"
)

func Regexp(expected any) Matcher {
	var re *regexp.Regexp

	if r, ok := expected.(*regexp.Regexp); ok {
		re = r
	} else {
		expected = fmt.Sprint(expected)
		if r, err := regexp.Compile(expected.(string)); err != nil {
			return matcherErr(fmt.Sprintf("Invalid Regexp(`%s`): %s", expected, err.Error()))
		} else {
			re = r
		}
	}

	return regexpMatcher{regexp: re}
}

type regexpMatcher struct {
	regexp *regexp.Regexp
}

func (m regexpMatcher) Matches(actual any) (ok bool, desc string) {
	if m.regexp.MatchString(fmt.Sprint(actual)) {
		return true, ""
	}

	return false, fmt.Sprintf("%s: not matched: %s", m.String(), formatValue(actual))
}

func (m regexpMatcher) String() string {
	return fmt.Sprintf("Regexp(`%s`)", m.regexp.String())
}
