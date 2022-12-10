package match_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grongor/go-muchtest"
	"github.com/grongor/go-muchtest/internal"
	"github.com/grongor/go-muchtest/match"
)

func TestEqualSuite(t *testing.T) {
	muchtest.Run(t, new(EqualSuite))
}

type EqualSuite struct {
	pkgSuite
}

type stringer struct {
	Text string
}

type unexported struct {
	number   int
	string   string
	stringer *stringer
	complex  complex128
}

func (s stringer) String() string {
	return s.Text
}

func (s *EqualSuite) TestEqual_Numbers() {
	type testT struct {
		name           string
		value1, value2 any
		ok             bool
	}

	values := []any{
		int(123), int8(123), int16(123), int32(123), int64(123),
		uint(123), uint8(123), uint16(123), uint32(123), uint64(123),
		float32(123.0), float64(123.0),
	}
	stringValues := []any{"123", stringer{Text: "123"}}
	differentValues := []any{
		int(124), int8(125), int16(126), int32(127), int64(128),
		uint(129), uint8(130), uint16(131), uint32(132), uint64(133),
		float32(123.1), float64(123.2),
	}
	differentStringValues := []any{"134", stringer{Text: "135"}}

	var tests []testT

	for _, value1 := range values {
		for _, subValues := range [][]any{values, stringValues} {
			for _, value2 := range subValues {
				tests = append(tests, testT{
					name:   fmt.Sprintf("equal %T and %T", value1, value2),
					value1: value1,
					value2: value2,
					ok:     true,
				})
			}
		}

		for _, subValues := range [][]any{differentValues, differentStringValues} {
			for _, value2 := range subValues {
				tests = append(tests, testT{
					name:   fmt.Sprintf("not equal %T and %T", value1, value2),
					value1: value1,
					value2: value2,
				})
			}
		}
	}

	errRe := regexp.MustCompile(`^Equal\(.*?\): not equal:`)

	for _, test := range tests {
		s.Run(test.name, func() {
			if test.ok {
				s.match(match.Equal(test.value1), test.value2)

				return
			}

			s.matchFn(match.Equal(test.value1, cmpopts.IgnoreUnexported(strings.Builder{})), test.value2, func(desc string) {
				s.S.Regexp(errRe, internal.TrimDiff(desc))
			})
		})
	}
}

func (s *EqualSuite) TestEqual() {
	type otherString string

	for _, test := range []struct {
		name string
		a, b any
		desc string
	}{
		// matching
		{name: "equal strings", a: "Much", b: "Much"},
		{name: "equal string and stringer", a: "Much", b: stringer{Text: "Much"}},
		{name: "equal strings, diff types", a: any("Much"), b: otherString("Much")},
		{name: "equal ints", a: 123, b: 123},
		{name: "equal ints, diff types", a: 123, b: uint8(123)},
		{name: "equal uints", a: uint(123), b: uint(123)},
		{name: "equal uints, diff types", a: uint(123), b: uint16(123)},
		{name: "equal floats", a: 12.3, b: 12.3},
		{name: "equal floats, diff types", a: float32(12.3), b: float64(12.3)},
		{name: "equal complex", a: complex(12.3, 13.4), b: complex(12.3, 13.4)},
		{
			name: "equal complex, diff types",
			a:    complex(float32(12.3), float32(13.4)),
			b:    complex(float64(12.3), float64(13.4)),
		},
		{name: "equal slice", a: []int{123, 456}, b: []int{123, 456}},
		{name: "equal slice, diff types", a: []int{123, 456}, b: []any{int16(123), int16(456)}},
		{
			name: "equal slice, recursive",
			a:    []any{123, float32(12.3), []any{[]string{"1", "2"}, map[int]uint{1: 2, 3: 4}}},
			b:    []any{int16(123), float64(12.3), []any{[]otherString{"1", "2"}, map[int8]uint16{1: 2, 3: 4}}},
		},
		{name: "equal map", a: map[string]int{"Much": 123, "Test": 123}, b: map[string]int{"Much": 123, "Test": 123}},
		{name: "equal map, diff types", a: map[string]int{"Much": 123, "Test": 123}, b: map[otherString]any{"Much": int16(123), "Test": int16(123)}},
		{
			name: "equal map, recursive",
			a: map[any]any{
				"Much": 123,
				2: map[int]any{
					3: float32(12.3),
					4: []float32{13.4, 14.5},
					5: "Much",
				},
			},
			b: map[any]any{
				otherString("Much"): int16(123),
				int8(2): map[int16]any{
					3: float64(12.3),
					4: []float64{13.4, 14.5},
					5: otherString("Much"),
				},
			},
		},
		{
			name: "equal struct",
			a: unexported{
				number:   123,
				string:   "Much",
				stringer: &stringer{Text: "Much"},
				complex:  123 + 4i,
			},
			b: unexported{
				number:   123,
				string:   "Much",
				stringer: &stringer{Text: "Much"},
				complex:  123 + 4i,
			},
		},

		// not matching
		{name: "invalid equal", a: "Much", b: nil, desc: `Equal("Much"): not equal: nil`},
		{name: "invalid actual", a: nil, b: "Much", desc: `Equal(nil): not equal: "Much"`},
		{name: "not equal strings", a: "Much", b: "Such", desc: `Equal("Much"): not equal: "Such"`},
		{
			name: "not equal strings, diff types",
			a:    "Much",
			b:    otherString("Such"),
			desc: `Equal("Much"): not equal: "Such"`,
		},
		{
			name: "not equal string and stringer",
			a:    "Much",
			b:    stringer{Text: "Such"},
			desc: `Equal("Much"): not equal: match_test.stringer(Such)`,
		},
		{name: "not equal ints", a: 123, b: 456, desc: `Equal(123): not equal: 456`},
		{name: "not equal ints, diff types", a: 123, b: int16(456), desc: `Equal(123): not equal: 456`},
		{name: "not equal uints", a: uint(123), b: uint(456), desc: `Equal(123): not equal: 456`},
		{name: "not equal uints, diff types", a: uint(123), b: uint16(456), desc: `Equal(123): not equal: 456`},
		{name: "not equal floats", a: 12.3, b: 13.4, desc: `Equal(12.3): not equal: 13.4`},
		{
			name: "not equal floats, diff types",
			a:    float32(12.3),
			b:    float64(13.4),
			desc: `Equal(12.3): not equal: 13.4`,
		},
		{
			name: "not equal complex",
			a:    complex(12.3, 13.4),
			b:    complex(12.3, 13.5),
			desc: `Equal((12.3+13.4i)): not equal: (12.3+13.5i)`,
		},
		{
			name: "not equal complex, diff types",
			a:    complex(float32(12.3), float32(13.4)),
			b:    complex(float64(12.3), float64(13.5)),
			desc: `Equal((12.3+13.4i)): not equal: (12.3+13.5i)`,
		},
		{
			name: "not equal slice",
			a:    []int{123, 456},
			b:    []int{456, 123},
			desc: `Equal([]int{123, 456}): not equal: []int{456, 123}`,
		},
		{
			name: "not equal slice, diff types",
			a:    []int{123, 456},
			b:    []any{int16(456), int16(123)},
			desc: `Equal([]int{123, 456}): not equal: []any{456, 123}`,
		},
		{
			name: "not equal slice, recursive",
			a:    []any{123, float32(12.3), []any{[]string{"1", "2"}, map[int]uint{1: 2, 3: 4}}},
			b:    []any{int16(123), float64(12.3), []any{[]otherString{"2", "1"}, map[any]uint16{any(1): 2, int8(3): 4}}},
			desc: `Equal([]any{123, 12.3, []any{[]string{"1", "2"}, map[int]uint{1:2, 3:4}}}): not` +
				` equal: []any{123, 12.3, []any{[]match_test.otherString{"2", "1"}, map[any]uint16{1:2, 3:4}}}`,
		},
		{
			name: "not equal map",
			a:    map[string]int{"Much": 123, "Test": 123},
			b:    map[string]int{"Much": 123, "Test": 1},
			desc: `Equal(map[string]int{"Much":123, "Test":123}): not equal: map[string]int{"Much":123, "Test":1}`,
		},
		{
			name: "not equal map, diff types",
			a:    map[string]int{"Much": 123, "Test": 123},
			b:    map[otherString]any{"Much": int16(123), "Test": int16(456)},
			desc: `Equal(map[string]int{"Much":123, "Test":123}): not` +
				` equal: map[match_test.otherString]any{"Much":123, "Test":456}`,
		},
		{
			name: "not equal map, recursive",
			a: map[any]any{
				"Much": 123,
				2: map[uint]any{
					3: float32(12.3),
					4: []float32{13.4, 14.5},
					5: "Much"}},
			b: map[any]any{
				otherString("Much"): int16(123),
				int8(2): map[uint16]any{
					3: float64(12.3),
					4: []float64{13.4, 14.6},
					5: otherString("Much"),
				},
			},
			desc: `Equal(map[any]any{"Much":123, 2:map[uint]any{3:12.3, 4:[]float32{13.4, 14.5}, 5:"Much"}}): not` +
				` equal: map[any]any{"Much":123, 2:map[uint16]any{3:12.3, 4:[]float64{13.4, 14.6}, 5:"Much"}}`,
		},
		{
			name: "not equal struct",
			a: unexported{
				number:   123,
				string:   "Much",
				stringer: &stringer{Text: "Much"},
				complex:  123 + 4i,
			},
			b: unexported{
				number:   123,
				string:   "Much",
				stringer: &stringer{Text: "Such"},
				complex:  123 + 4i,
			},
			desc: "Equal(match_test.unexported{}): not equal: match_test.unexported{}",
		},
	} {
		s.Run(test.name, func() {
			s.match(match.Equal(test.a, cmp.AllowUnexported(unexported{})), test.b, test.desc)
		})
	}
}
