package match

type SimpleMatcher interface {
	Matches(actual any) bool
}

type Matcher interface {
	Matches(actual any) (ok bool, desc string)
	String() string
}

type StatefulMatcher interface {
	Matcher
	Stateful() Matcher
}

func Any() Matcher {
	return nopMatcher{}
}

func None() Matcher {
	return nopMatcher{desc: "None() never matches"}
}

type nopMatcher struct {
	desc string
}

func (m nopMatcher) Matches(any) (ok bool, desc string) {
	if m.desc == "" {
		return true, ""
	}

	return false, m.desc
}

func (m nopMatcher) String() string {
	if m.desc == "" {
		return "Any()"
	}

	return "None()"
}
