package match_test

import (
	"regexp"
	"testing"

	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/match"
)

func TestRegexpSuite(t *testing.T) {
	muchtest.Run(t, new(RegexpSuite))
}

type RegexpSuite struct {
	pkgSuite
}

func (s *RegexpSuite) TestRegexp() {
	const matchingRegexp = `.*est\W+Mu\w+\?`
	const notMatchingRegexp = `\w+ not Much?`

	s.S.Equal("Regexp(`Re.*x`)", match.Regexp(`Re.*x`).String())
	s.S.Equal("Regexp(`Re.*x`)", match.Regexp(regexp.MustCompile(`Re.*x`)).String())

	s.match(match.Regexp(matchingRegexp), "Test Much?")
	s.match(match.Regexp(regexp.MustCompile(matchingRegexp)), "Test Much?")
	s.match(match.Regexp(struct{ Hello string }{Hello: "wow"}), "Can I hear a {wow}?")
	s.match(match.Regexp(notMatchingRegexp), "Test Much?", "Regexp(`\\w+ not Much?`): not matched: \"Test Much?\"")
	s.match(match.Regexp(regexp.MustCompile(notMatchingRegexp)), "Test Much?",
		"Regexp(`\\w+ not Much?`): not matched: \"Test Much?\"")
	s.match(match.Regexp(struct{ Hello string }{Hello: "wow"}), "Can I hear a WOW?!",
		"Regexp(`{wow}`): not matched: \"Can I hear a WOW?!\"")
	s.match(match.Regexp(`*invalid`), "Test Much?",
		"Invalid Regexp(`*invalid`): error parsing regexp: missing argument to repetition operator: `*`")
}
