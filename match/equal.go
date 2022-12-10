package match

import (
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
)

func Equal(expected any, options ...cmp.Option) Matcher {
	return equalMatcher{equal: expected, options: options}
}

type noStringer struct {
	v any
}

func (s noStringer) String() string {
	return fmt.Sprint(s.v)
}

type equalMatcher struct {
	equal   any
	options cmp.Options
}

func (m equalMatcher) Matches(actual any) (ok bool, desc string) {
	defer func() {
		if recover() == nil && ok {
			return
		}

		ok = false
		desc = fmt.Sprintf("%s: not equal: %s%s", m.String(), formatValue(actual), getDiff(m.equal, actual, m.options))
	}()

	ok = m.doMatches(reflectV(m.equal), reflectV(actual))

	return
}

func (m equalMatcher) doMatches(vEqual, vActual reflect.Value) bool {
	vEqual = indirect(vEqual)
	vActual = indirect(vActual)

	if !vEqual.IsValid() || !vActual.IsValid() {
		return false
	}

	if vActual.Kind() == reflect.Pointer {
		return m.doMatches(vEqual, vActual.Elem())
	}

	switch vEqual.Kind() {
	case reflect.String:
		if vActual.Kind() != reflect.String {
			if vActual.CanInterface() {
				return vEqual.String() == fmt.Sprint(vActual.Interface())
			}

			if vActual.Type().Implements(reflectT(new(fmt.Stringer)).Elem()) {
				stringResult := vActual.MethodByName("String").Call(nil)

				return fmt.Sprint(vEqual.Interface()) == stringResult[0].Interface().(string)
			}

			return false
		}

		return vEqual.String() == vActual.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		switch vActual.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64:
			ok, _ := Between(vEqual, vEqual).Matches(vActual)

			return ok
		case reflect.String:
			a := fmt.Sprint(vEqual.Interface())
			b := vActual.String()
			eq := a == b
			return eq
		default:
			if vActual.Type().Implements(reflectT(new(fmt.Stringer)).Elem()) {
				if vActual.CanInterface() {
					return fmt.Sprint(noStringer{v: vEqual.Interface()}) == fmt.Sprint(vActual.Interface())
				}

				stringResult := vActual.MethodByName("String").Call(nil)

				return fmt.Sprint(noStringer{v: vEqual.Interface()}) == stringResult[0].Interface().(string)
			}

			return false
		}
	case reflect.Complex64, reflect.Complex128:
		if !vActual.CanComplex() {
			return false
		}

		if vEqual.CanInterface() && vActual.CanInterface() {
			return fmt.Sprint(vEqual.Interface()) == fmt.Sprint(vActual.Interface())
		}

		return fmt.Sprint(vEqual.Complex()) == fmt.Sprint(vActual.Complex())
	case reflect.Slice, reflect.Array:
		return m.matchSlice(vEqual, vActual)
	case reflect.Map:
		return m.matchMap(vEqual, vActual)
	case reflect.Struct:
		return m.matchStruct(vEqual, vActual)
	case reflect.Pointer:
		if vActual.Kind() == reflect.Pointer {
			return m.doMatches(vEqual.Elem(), vActual.Elem())
		}

		return m.doMatches(vEqual.Elem(), vActual)
	default:
		return cmp.Equal(vEqual.Interface(), vActual.Interface(), m.options...)
	}
}

func (m equalMatcher) matchSlice(vEqual, vActual reflect.Value) bool {
	expectedLength := vEqual.Len()
	if vActual.Kind() != reflect.Slice || expectedLength != vActual.Len() {
		return false
	}

	for i := 0; i < expectedLength; i++ {
		if !m.doMatches(vEqual.Index(i), vActual.Index(i)) {
			return false
		}
	}

	return true
}

func (m equalMatcher) getMapValue(vEqualKey, vActual reflect.Value) (vActualValue reflect.Value) {
	defer func() {
		if recover() == nil && vActualValue.IsValid() {
			return
		}

		for _, vActualKey := range vActual.MapKeys() {
			if m.doMatches(vEqualKey, vActualKey) {
				vActualValue = vActual.MapIndex(vActualKey)

				return
			}
		}
	}()

	return vActual.MapIndex(vEqualKey)
}

func (m equalMatcher) matchMap(vEqual, vActual reflect.Value) bool {
	if vActual.Kind() != reflect.Map || vEqual.Len() != vActual.Len() {
		return false
	}

	vEqualKeys := vEqual.MapKeys()

	if len(vEqualKeys) != vActual.Len() {
		return false
	} else if len(vEqualKeys) == 0 {
		return true
	}

	for _, vEqualKey := range vEqualKeys {
		if !m.doMatches(vEqual.MapIndex(vEqualKey), m.getMapValue(vEqualKey, vActual)) {
			return false
		}
	}

	return true
}

func (m equalMatcher) matchStruct(vEqual, vActual reflect.Value) bool {
	fieldsCount := vEqual.NumField()
	if vActual.Kind() != reflect.Struct || fieldsCount != vActual.NumField() {
		return false
	}

	tEqual := vEqual.Type()
	tActual := vActual.Type()

	for i := 0; i < fieldsCount; i++ {
		if tEqual.Field(i).Name != tActual.Field(i).Name {
			return false
		}

		if !m.doMatches(vEqual.Field(i), vActual.Field(i)) {
			return false
		}
	}

	return true
}

func (m equalMatcher) String() string {
	return fmt.Sprintf("Equal(%s)", formatValue(m.equal))
}
