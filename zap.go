package muchtest

import (
	"fmt"
	"strings"

	"github.com/grongor/go-muchtest/match"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type zapState int

const (
	zapCreated = zapState(iota)
	zapInitialized
	zapAsserting
	zapAsserted

	zapPrefix = "zap: "
)

func NewZap(t TestingT) *Zap {
	return &Zap{t: t}
}

type Zap struct {
	t                         TestingT
	state                     zapState
	logger                    *zap.Logger
	partial, orderless, loose bool

	observedLogs *observer.ObservedLogs
	index        int
	usedIndexes  []int
	pending      *zapEntryAssert
}

func (z *Zap) Get() *zap.Logger {
	return z.GetAt(zap.InfoLevel)
}

func (z *Zap) GetAt(level zapcore.Level) *zap.Logger {
	z.t.Helper()
	z.transition(zapInitialized)

	var core zapcore.Core
	core, z.observedLogs = observer.New(level)

	z.logger = zap.New(core, zap.WithFatalHook(zapcore.WriteThenGoexit))

	return z.logger
}

func (z *Zap) Use(logger *zap.Logger, observedLogs *observer.ObservedLogs) *Zap {
	z.t.Helper()
	z.transition(zapInitialized)

	z.logger = logger
	z.observedLogs = observedLogs

	return z
}

func (z *Zap) ObservedLogs() *observer.ObservedLogs {
	z.t.Helper()
	z.transition(zapAsserted)

	return z.observedLogs
}

func (z *Zap) Debug(message ...any) *zapEntryAssert {
	return z.doNewAssert(nil).setLevelAndMessage(zapcore.DebugLevel, message)
}

func (z *Zap) Info(message ...any) *zapEntryAssert {
	return z.doNewAssert(nil).setLevelAndMessage(zapcore.InfoLevel, message)
}

func (z *Zap) Warn(message ...any) *zapEntryAssert {
	return z.doNewAssert(nil).setLevelAndMessage(zapcore.WarnLevel, message)
}

func (z *Zap) Error(message ...any) *zapEntryAssert {
	return z.doNewAssert(nil).setLevelAndMessage(zapcore.ErrorLevel, message)
}

func (z *Zap) Panic(message ...any) *zapEntryAssert {
	return z.doNewAssert(nil).setLevelAndMessage(zapcore.PanicLevel, message)
}

func (z *Zap) Fatal(message ...any) *zapEntryAssert {
	return z.doNewAssert(nil).setLevelAndMessage(zapcore.FatalLevel, message)
}

func (z *Zap) Assert(name string) *zapEntryAssert {
	return z.doNewAssert(&name)
}

func (z *Zap) doNewAssert(name *string) *zapEntryAssert {
	z.t.Helper()
	z.transition(zapAsserting)
	z.checkPendingAssert()

	z.pending = &zapEntryAssert{z: z, name: name}

	return z.pending
}

func (z *Zap) checkPendingAssert() {
	if z.pending == nil {
		return
	}

	const message = "First call one of the terminating methods(%s) on the previous assert: %s"
	z.reportError(fmt.Sprintf(message, zapEntryAssertTerminatingMethods, z.pending.String()))
}

func (z *Zap) AssertNext(level zapcore.Level, message any, ctxKeyAndValues ...any) *Zap {
	z.t.Helper()

	return z.doAssertNext(nil, level, message, ctxKeyAndValues)
}

func (z *Zap) AssertNamed(name string, level zapcore.Level, message any, ctxKeyAndValues ...any) *Zap {
	z.t.Helper()

	return z.doAssertNext(&name, level, message, ctxKeyAndValues)
}

func (z *Zap) SkipNext(n ...int) *Zap {
	z.t.Helper()
	z.transition(zapAsserting)

	if l := len(n); l > 0 {
		if l != 1 || n[0] < 1 {
			z.reportError("Parameter to SkipNext() must be a single positive integer, or nothing (equivalent to 1)")
		}

		z.index += n[0]
	} else {
		z.index++
	}

	return z
}

func (z *Zap) AssertNoNextEntry() {
	z.t.Helper()
	z.transition(zapAsserting)

	if message := z.checkNoNextEntry(); message != "" {
		z.reportError(message)
	}
}

func (z *Zap) checkNoNextEntry() string {
	l := z.observedLogs.Len()
	if l == z.index {
		return ""
	}

	builder := &strings.Builder{}

	_, _ = fmt.Fprintf(builder, "AssertNoNextEntry(): there are entries available (%d)\n", l-z.index)

	for _, entry := range z.observedLogs.All()[z.index:] {
		builder.WriteByte('\t')
		z.formatEntry(builder, entry)
		builder.WriteByte('\n')
	}

	return builder.String()
}

func (z *Zap) formatEntry(builder *strings.Builder, entry observer.LoggedEntry) {
	builder.WriteByte('[')
	builder.WriteString(entry.Time.UTC().Format("2006-01-02 15:04:05.999999"))
	builder.WriteString("] ")
	builder.WriteString(entry.Level.CapitalString())
	builder.WriteString(" ")

	if entry.LoggerName != "" {
		builder.WriteByte('[')
		builder.WriteString(entry.LoggerName)
		builder.WriteString("] ")
	}

	builder.WriteString(entry.Message)

	contextLen := len(entry.Context)
	if contextLen != 0 {
		builder.WriteString("\n\t\t[")

		for k, v := range entry.ContextMap() {
			fmt.Fprintf(builder, `"%s": "%s"`, k, v)
			contextLen--

			if contextLen != 0 {
				builder.WriteString(", ")
			}
		}

		builder.WriteByte(']')
	}
}

func (z *Zap) IgnoreMissingContext(partial ...bool) *Zap {
	z.t.Helper()

	z.partial = z.getBool(partial, "IgnoreMissingContext")

	return z
}

func (z *Zap) Orderless(orderless ...bool) *Zap {
	z.t.Helper()

	newOrderless := z.getBool(orderless, "Orderless")

	if z.orderless {
		if newOrderless {
			z.reportError("Orderless(): already set to orderless")
		}

		if len(z.usedIndexes) != 0 {
			z.index += len(z.usedIndexes)
			z.usedIndexes = nil

			if z.index != z.observedLogs.Len() {
				z.reportError("Orderless(): some messages weren't asserted")
			}
		}
	} else if !newOrderless {
		z.reportError("Orderless(): already set to ordered")
	}

	z.orderless = newOrderless

	return z
}

func (z *Zap) MayContainMoreEntries() {
	z.loose = true
}

func (z *Zap) transition(state zapState) {
	if z.state == zapCreated && state != zapInitialized {
		z.reportError("First call Get(), GetAt() or Use() methods")
	}

	if z.state >= zapInitialized && state == zapInitialized {
		z.reportError("You can't call Get(), GetAt() or Use() multiple times")
	}

	if z.state == zapAsserting && state == zapAsserted || z.state == zapAsserted && state == zapAsserting {
		z.reportError("You can't combine ObservedLogs() and AssertNext*() methods")
	}

	if z.state == zapAsserted && state == zapAsserted {
		z.reportError("You can't call ObservedLogs() multiple times")
	}

	z.state = state
}

func (z *Zap) doAssertNext(name *string, level zapcore.Level, message any, ctxKeyAndValues []any) *Zap {
	if len(ctxKeyAndValues) == 1 {
		if matcher, ok := ctxKeyAndValues[0].(match.Matcher); ok {
			return z.doAssert(name, level, message, matcher)
		}
	}

	return z.doAssert(name, level, message, match.Fn(func(ctxMap map[string]any) (bool, string) {
		if len(ctxKeyAndValues)%2 != 0 {
			n := "AssertNext"
			if name != nil {
				n = "AssertNamed"
			}

			return false, n + `(): ctxKeyAndValues must be pairs; eg.: "duration", 5*time.Second, "status", "ok"`
		}

		if !z.partial {
			if ok, desc := match.Len(len(ctxKeyAndValues) / 2).Matches(ctxMap); !ok {
				return ok, desc
			}
		}

		for i := 1; i < len(ctxKeyAndValues); i += 2 {
			if ok, desc := match.ContainsKeyValue(ctxKeyAndValues[i-1], ctxKeyAndValues[i]).Matches(ctxMap); !ok {
				return ok, desc
			}
		}

		return true, ""
	}))
}

func (z *Zap) doAssert(name *string, level zapcore.Level, message any, ctxMatcher match.Matcher) *Zap {
	z.transition(zapAsserting)

	if z.orderless {
	Remaining:
		for i := z.index; i < z.observedLogs.Len(); i++ {
			for _, j := range z.usedIndexes {
				if i == j {
					continue Remaining
				}
			}

			entry := z.observedLogs.All()[i]

			if ok, _ := z.checkEntry(entry, name, level, message, ctxMatcher); ok {
				z.usedIndexes = append(z.usedIndexes, i)

				return z
			}
		}

		if z.loose {
			return z
		}

		z.reportError("No entries matched and there are no more logged entries")
	}

	if z.index == z.observedLogs.Len() {
		if z.loose {
			return z
		}

		z.reportError("There are no more logged entries")
	}

	entry := z.observedLogs.All()[z.index]
	z.index++

	if ok, desc := z.checkEntry(entry, name, level, message, ctxMatcher); !ok {
		z.reportError(strings.TrimPrefix(desc, "Fn(): "))
	}

	return z
}

func (z *Zap) checkEntry(
	entry observer.LoggedEntry,
	name *string,
	level zapcore.Level,
	message any,
	ctxMatcher match.Matcher,
) (bool, string) {
	if name != nil {
		if ok, desc := match.Equal(*name).Matches(entry.LoggerName); !ok {
			return ok, desc
		}
	}

	if ok, desc := match.Equal(level).Matches(entry.Level); !ok {
		return ok, desc
	}

	if ok, desc := match.ToMatcher(message).Matches(entry.Message); !ok {
		return ok, desc
	}

	if ok, desc := ctxMatcher.Matches(entry.ContextMap()); !ok {
		desc += "\n"

		return ok, desc
	}

	return true, ""
}

func (z *Zap) getBool(b []bool, caller string) bool {
	switch len(b) {
	case 0:
		return true
	case 1:
		return b[0]
	default:
		z.reportError(caller + "(): parameter must be a single bool, or nothing (equivalent to true)")

		return false
	}
}

type OrderlessZap interface {
}

func (z *Zap) reportError(err string) {
	require.Fail(z.t, prefix+zapPrefix+err)
	//z.assert.Fail("muchtest: zap: " + err)
	//z.t.Errorf("muchtest: zap: %s", err)
	//z.t.FailNow()
}
