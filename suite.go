package muchtest

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const prefix = "muchtest: "

type TestingSuite interface {
	suite.TestingSuite

	SetSelf(self TestingSuite) TestingSuite
}

type Suite struct {
	suite.Suite

	S *ourSuite // @see https://youtrack.jetbrains.com/issue/GO-13592/Cannot-run-tests-in-custom-testify-suite
}

func (s *Suite) SetSelf(self TestingSuite) TestingSuite {
	s.S = &ourSuite{
		self:  self,
		vSelf: reflect.ValueOf(self).Elem(),
	}

	return self
}

func (s *Suite) T() *testing.T {
	return s.S.t
}

func (s *Suite) SetT(t *testing.T) {
	s.S.SetT(t)
}

func (s *Suite) SetupSuite() {
	s.S.SetupSuite()
}

func (s *Suite) SetupTest() {
	s.S.SetupTest()
}

func (s *Suite) TearDownTest() {
	s.S.TearDownTest()
}

func (s *Suite) Run(name string, subtest func()) bool {
	return s.S.Run(name, subtest)
}

type ourSuite struct {
	suite.TestingSuite
	Assertions

	self  TestingSuite
	vSelf reflect.Value
	t     *testing.T
	z     *Zap
	clock clockwork.FakeClock
	mocks []reflect.Value
	mu    sync.Mutex
}

func (s *ourSuite) Zap() *Zap {
	return s.z
}

func (s *ourSuite) Clock() clockwork.FakeClock {
	if s.clock == nil {
		s.clock = clockwork.NewFakeClock()
	}

	return s.clock
}

func (s *ourSuite) ClockAt(t time.Time) clockwork.FakeClock {
	if s.clock != nil {
		require.Fail(s.t, "You can't call ClackAt() multiple times, or after calling Clock()")
	}

	s.clock = clockwork.NewFakeClockAt(t)

	return s.clock
}

func (s *ourSuite) Run(name string, fn func()) bool {
	oldT := s.T()
	defer s.SetT(oldT)

	return s.T().Run(name, func(t *testing.T) {
		s.SetT(t)
		s.SetupTest()

		if before, ok := s.self.(suite.BeforeTest); ok {
			names := strings.SplitN(t.Name(), "/", 2)

			before.BeforeTest(strings.TrimPrefix(names[0], "Test"), names[1])
		}

		defer s.self.(suite.TearDownTestSuite).TearDownTest()
		defer func() {
			if after, ok := s.self.(suite.AfterTest); ok {
				names := strings.SplitN(t.Name(), "/", 2)

				after.AfterTest(strings.TrimPrefix(names[0], "Test"), names[1])
			}
		}()

		fn()
	})
}

func (s *ourSuite) T() *testing.T {
	return s.t
}

func (s *ourSuite) SetT(t *testing.T) {
	s.t = t
	s.Assertions.t = t
	s.z = NewZap(t)
}

func (s *ourSuite) SetupSuite() {
	fields := s.vSelf.NumField()

	for i := 0; i < fields; i++ {
		field := s.vSelf.Field(i)
		if field.Kind() != reflect.Ptr || !field.CanSet() {
			continue
		}

		fieldType := field.Type().Elem()

		if mockField, ok := fieldType.FieldByName("Mock"); !ok || mockField.Type != reflect.TypeOf(mock.Mock{}) {
			continue
		}

		s.mocks = append(s.mocks, field)
	}
}

func (s *ourSuite) SetupTest() {
	s.mu.Lock()
	defer s.mu.Unlock()

	argsTestingT := []reflect.Value{reflect.ValueOf(s.t)}

	for _, field := range s.mocks {
		field.Set(reflect.New(field.Type().Elem()))
		field.MethodByName("Test").Call(argsTestingT)
	}
}

func (s *ourSuite) TearDownTest() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clock = nil

	argsTestingT := []reflect.Value{reflect.ValueOf(s.t)}

	for _, field := range s.mocks {
		field.MethodByName("AssertExpectations").Call(argsTestingT)
	}

	if s.z.logger == nil {
		return
	}

	if s.z.pending != nil {
		assert.Fail(s.t, fmt.Sprintf("%s%sMissing call to one of the terminating methods(%s) on assert: %s",
			prefix, zapPrefix, zapEntryAssertTerminatingMethods, s.z.pending.String()))
	}

	if !s.z.loose {
		if message := s.z.checkNoNextEntry(); message != "" {
			assert.Fail(s.t, prefix+zapPrefix+message)
		}
	}
}

func Run(t *testing.T, s TestingSuite) {
	t.Parallel()
	suite.Run(t, s.SetSelf(s))
}
