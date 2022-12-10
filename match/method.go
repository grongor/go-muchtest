package match

import (
	"fmt"
	"reflect"
)

func Method(method string, actual ...any) Matcher {
	return methodMatcher{method: method, matchers: ToMatchers(actual)}
}

type methodMatcher struct {
	method   string
	args     []any
	matchers []Matcher
}

func (m methodMatcher) Matches(actual any) (ok bool, desc string) {
	vActual := reflectV(actual)

	vMethod := vActual.MethodByName(m.method)
	if !vMethod.IsValid() {
		return false, fmt.Sprintf(`Method(%s): not defined on: %s`, m.method, formatValue(actual))
	}

	tMethod := vMethod.Type()
	returnValuesCount := tMethod.NumOut()
	matchersCount := len(m.matchers)

	if returnValuesCount != matchersCount {
		return false, fmt.Sprintf(
			`Method(%s): expected %d return values, got %d, value: %s`,
			m.method, returnValuesCount, matchersCount, formatValue(actual),
		)
	}

	var args []reflect.Value

	for _, arg := range m.args {
		args = append(args, reflectV(arg))
	}

	defer func() {
		if r := recover(); r != nil {
			ok = false
			desc = fmt.Sprint("duh fail", r)
		}
	}()

	result := vMethod.Call(args)

	for i := 0; i < matchersCount; i++ {
		if ok, desc = m.matchers[i].Matches(result[i].Interface()); !ok {
			return false, fmt.Sprintf(`Method(%s): return value at index %d not matched: %s`, m.method, i, desc)
		}
	}

	return true, ""
}

func (m methodMatcher) String() string {
	return fmt.Sprintf("Method(%s)", m.method)
}
