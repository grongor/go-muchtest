package match

import (
	"bufio"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
	"github.com/grongor/go-muchtest/internal"
)

func ToMatchers(expected []any) []Matcher {
	matchers := make([]Matcher, len(expected))

	for i, e := range expected {
		matchers[i] = ToMatcher(e)
	}

	return matchers
}

func ToMatcher(expected any) Matcher {
	if matcher, ok := expected.(Matcher); ok {
		return matcher
	}

	if re, ok := expected.(*regexp.Regexp); ok {
		return Regexp(re)
	}

	if reflectT(expected).Kind() == reflect.Func {
		return Fn(expected)
	}

	return Equal(expected)
}

type matcherErrMatcher struct {
	err    string
	params []any
}

func matcherErr(err string, params ...any) matcherErrMatcher {
	return matcherErrMatcher{err: err, params: params}
}

func (m matcherErrMatcher) Matches(any) (ok bool, desc string) {
	if len(m.params) == 0 {
		return false, m.err
	}

	return false, fmt.Sprintf("%s Parameters: %+v", m.err, m.params)
}

func (m matcherErrMatcher) Stateful() Matcher {
	return m
}

func (m matcherErrMatcher) String() string {
	return "{InvalidMatcher}"
}

func indirect(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Interface {
		return value.Elem()
	}

	return value
}

func reflectV(value any) reflect.Value {
	if v, ok := value.(reflect.Value); ok {
		return v
	}

	return reflect.ValueOf(value)
}

func reflectT(value any) reflect.Type {
	if t, ok := value.(reflect.Type); ok {
		return t
	}

	if v, ok := value.(reflect.Value); ok && v.IsValid() {
		return v.Type()
	}

	return reflect.TypeOf(value)
}

func stringAny(s string) string {
	return strings.ReplaceAll(s, "interface {}", "any")
}

func formatValue(value any) string {
	builder := strings.Builder{}

	doFormatValue(&builder, value)

	return builder.String()
}

func doFormatValue(builder *strings.Builder, value any) {
	tValue := reflectT(value)
	t2 := reflectT(reflect.Value{})
	if reflect.TypeOf(value) == t2 {
		value = value.(reflect.Value).Interface()
		tValue = reflectT(value)
	}

	if s, ok := formatString(value); ok {
		builder.WriteString(s)

		return
	}

	if tValue == nil {
		builder.WriteString("nil")

		return
	}

	switch tValue.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fmt.Fprint(builder, value)
	case reflect.Slice, reflect.Array:
		vValue := reflectV(value)
		valueLen := vValue.Len()

		sliceValueType := stringAny(tValue.Elem().String())

		builder.WriteString("[]")
		builder.WriteString(sliceValueType)
		builder.WriteByte('{')

		for i := 0; i < valueLen; i++ {
			if i != 0 {
				builder.WriteString(", ")
			}

			doFormatValue(builder, vValue.Index(i).Interface())
		}

		builder.WriteByte('}')
	case reflect.Map:
		vValue := reflectV(value)
		keyType, valueType := tValue.Key().String(), tValue.Elem().String()

		for _, x := range [...]*string{&keyType, &valueType} {
			*x = stringAny(*x)
		}

		builder.WriteString("map[")
		builder.WriteString(keyType)
		builder.WriteByte(']')
		builder.WriteString(valueType)
		builder.WriteByte('{')

		type keyAndString struct {
			vKey   reflect.Value
			string string
		}

		vKeys := vValue.MapKeys()
		keyAndStrings := make([]keyAndString, len(vKeys))

		for i, vKey := range vKeys {
			if s, ok := formatString(vKey.Interface()); ok {
				keyAndStrings[i] = keyAndString{vKey: vKey, string: s}

				continue
			}

			keyAndStrings[i] = keyAndString{vKey: vKey, string: fmt.Sprint(vKey.Interface())}
		}

		sort.SliceStable(keyAndStrings, func(i, j int) bool {
			return keyAndStrings[i].string < keyAndStrings[j].string
		})

		for i, ks := range keyAndStrings {
			if i != 0 {
				builder.WriteString(", ")
			}

			builder.WriteString(ks.string)
			builder.WriteByte(':')
			doFormatValue(builder, vValue.MapIndex(ks.vKey).Interface())
		}

		builder.WriteByte('}')
	case reflect.Struct:
		vValue := reflectV(value)

		builder.WriteString(vValue.Type().String())

		if vValue.IsZero() {
			builder.WriteString("{}")

			return
		}

		builder.WriteByte('{')

		type fieldAndName struct {
			field reflect.Value
			name  string
		}

		fieldsCount := vValue.NumField()
		fieldAndNames := make([]fieldAndName, 0, fieldsCount)

		for i := 0; i < fieldsCount; i++ {
			vField := vValue.Field(i)
			if !vField.CanInterface() || vField.IsZero() {
				continue
			}

			fieldAndNames = append(fieldAndNames, fieldAndName{field: vField, name: tValue.Field(i).Name})
		}

		sort.SliceStable(fieldAndNames, func(i, j int) bool {
			return fieldAndNames[i].name < fieldAndNames[j].name
		})

		for i, f := range fieldAndNames {
			if i != 0 {
				builder.WriteString(", ")
			}

			builder.WriteString(f.name)
			builder.WriteByte(':')
			doFormatValue(builder, f.field.Interface())
		}

		builder.WriteByte('}')
	case reflect.Pointer:
		vValue := reflectV(value)

		builder.WriteByte('*')

		if vValue.IsNil() {
			builder.WriteString(vValue.Type().String())
			builder.WriteString("(nil)")
		} else {
			doFormatValue(builder, vValue.Elem().Interface())
		}
	default:
		builder.WriteString(stringAny(fmt.Sprintf("%#v", value)))
	}

}

func toString(value any) (str string, ok bool) {
	return doToString(value, false)
}

func formatString(value any) (str string, ok bool) {
	return doToString(value, true)
}

func doToString(value any, format bool) (str string, ok bool) {
	if s, ok := value.(*regexp.Regexp); ok {
		return "`" + s.String() + "`", true
	}

	if t, ok := value.(time.Time); ok {
		return t.UTC().Format("2006-01-02 15:04:05.999999"), true
	}

	if s, ok := value.(Matcher); ok {
		return s.String(), true
	}

	if s, ok := value.(fmt.Stringer); ok {
		return fmt.Sprintf("%s(%s)", reflectV(value).Type().String(), s.String()), true
	}

	vValue := reflectV(value)

	switch vValue.Kind() {
	case reflect.String:
		if !format {
			return vValue.String(), true
		}

		return `"` + vValue.String() + `"`, true
	case reflect.Slice, reflect.Array:
		if vValue.Type().Elem().Kind() == reflect.Uint8 {
			if b := vValue.Bytes(); utf8.Valid(b) {
				if !format {
					return string(b), true
				}

				return `"` + string(b) + `"`, true
			}
		}
	}

	return "", false
}

func toStringTypes() string {
	return "string, fmt.Stringer, []byte, time.Time, *regexp.Regexp"
}

func getDiff(a, b any, options cmp.Options) string {
	diff := cmp.Diff(a, b, options...)
	reader := strings.NewReader(diff)
	scanner := bufio.NewScanner(reader)

	const linePrefix = "\t\t"

	diffBuilder := strings.Builder{}
	diffBuilder.Grow(len(internal.DiffPrefix) + reader.Len() +
		strings.Count(diff, "\n")*len(linePrefix) - len(linePrefix))
	diffBuilder.WriteString(internal.DiffPrefix)

	for scanner.Scan() {
		if len(scanner.Bytes()) != 0 {
			diffBuilder.WriteString(linePrefix)
			diffBuilder.Write(scanner.Bytes())
		}

		diffBuilder.WriteByte('\n')
	}
	if scanner.Err() != nil {
		diffBuilder.Reset()
		diffBuilder.WriteString(diff)
	}

	return diffBuilder.String()
}
