package muchtest

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/grongor/go-muchtest/match"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const zapEntryAssertTerminatingMethods = "Ctx, NoCtx, IgnoreCtx, CtxErr and IgnoreMissingCtx"

type zapEntryAssert struct {
	error // just a neat trick; IDE marks calls without termination methods

	z             *Zap
	name          *string
	level         *zapcore.Level
	message       any
	partial, done bool
}

func (a *zapEntryAssert) Debug(message ...any) *zapEntryAssert {
	return a.setLevelAndMessage(zapcore.DebugLevel, message)
}

func (a *zapEntryAssert) Info(message ...any) *zapEntryAssert {
	return a.setLevelAndMessage(zapcore.InfoLevel, message)
}

func (a *zapEntryAssert) Warn(message ...any) *zapEntryAssert {
	return a.setLevelAndMessage(zapcore.WarnLevel, message)
}

func (a *zapEntryAssert) Error(message ...any) *zapEntryAssert {
	return a.setLevelAndMessage(zapcore.ErrorLevel, message)
}

func (a *zapEntryAssert) Panic(message ...any) *zapEntryAssert {
	return a.setLevelAndMessage(zapcore.PanicLevel, message)
}

func (a *zapEntryAssert) Fatal(message ...any) *zapEntryAssert {
	return a.setLevelAndMessage(zapcore.FatalLevel, message)
}

func (a *zapEntryAssert) IgnoreMissingCtx(ctxKeyAndValues ...any) *Zap {
	a.partial = true

	return a.Ctx(ctxKeyAndValues)
}

func (a *zapEntryAssert) CtxErr(err any) *Zap {
	return a.IgnoreMissingCtx("error", err)
}

func (a *zapEntryAssert) Ctx(ctxKeyAndValues ...any) *Zap {
	if len(ctxKeyAndValues) == 1 {
		if matcher, ok := ctxKeyAndValues[0].(match.Matcher); ok {
			return a.doAssert(matcher)
		}
	}

	return a.doAssert(match.Fn(func(ctxMap map[string]any) (bool, string) {
		if len(ctxKeyAndValues)%2 != 0 {
			return false, "Ctx(): ctxKeyAndValues must be a matcher, or pairs like: " +
				`"duration", 5*time.Second, "status", "ok"`
		}

		if !a.partial && !a.z.partial {
			if ok, desc := match.Len(len(ctxKeyAndValues) / 2).Matches(ctxMap); !ok {
				return ok, desc
			}
		}

		for i := 0; i < len(ctxKeyAndValues); i += 2 {
			key := ctxKeyAndValues[i]
			var value any

			if field, ok := key.(zapcore.Field); ok {
				enc := zapcore.NewMapObjectEncoder()

				field.AddTo(enc)

				key = field.Key
				value = enc.Fields[field.Key]
				i--
			} else if i < len(ctxKeyAndValues) {
				value = ctxKeyAndValues[i+1]
			} else {
				return false, "Ctx(): ctxKeyAndValues must be a matcher, or (possible) combination of zapcore.Field(s)" +
					`and/or pairs like: "duration", 5*time.Second, "status", "ok"`
			}

			if ok, desc := match.ContainsKeyValue(key, value).Matches(ctxMap); !ok {
				return ok, desc
			}
		}

		return true, ""
	}))
}

func (a *zapEntryAssert) NoCtx() *Zap {
	return a.Ctx(match.Len(0))
}

func (a *zapEntryAssert) IgnoreCtx() *Zap {
	return a.doAssert(nil)
}

func (a *zapEntryAssert) String() string {
	builder := strings.Builder{}

	if a.name != nil {
		builder.WriteString(`Assert("`)
		builder.WriteString(*a.name)
		builder.WriteString(`")`)

		if a.level == nil && a.message == nil {
			return builder.String()
		}

		builder.WriteByte('.')
	}

	writeMessage := func() {
		if msg, ok := a.message.(string); ok {
			builder.WriteByte('"')
			builder.WriteString(msg)
			builder.WriteByte('"')
		} else {
			fmt.Fprintf(&builder, "%v", a.message)
		}
	}

	if a.level != nil {
		builder.WriteString(cases.Title(language.English, cases.NoLower).String(a.level.String()))
		builder.WriteByte('(')

		if a.message != nil {
			writeMessage()
		}

		builder.WriteByte(')')
	} else if a.message != nil {
		builder.WriteString("Msg(")
		writeMessage()
		builder.WriteByte(')')
	}

	return builder.String()
}

func (a *zapEntryAssert) setLevelAndMessage(level zapcore.Level, message []any) *zapEntryAssert {
	if a.level != nil {
		const message = "Can't call %s(): %s() already called"
		a.z.reportError(fmt.Sprintf(message,
			cases.Title(language.English, cases.NoLower).String(level.String()),
			cases.Title(language.English, cases.NoLower).String(a.level.String()),
		))
	}

	a.level = &level

	if len(message) != 0 {
		if len(message) > 1 || message[0] == nil {
			const message = "Message parameter to %s() must be a single string or matcher, or nothing"
			a.z.reportError(fmt.Sprintf(message, cases.Title(language.English, cases.NoLower).String(level.String())))
		}

		return a.setMessage(message[0])
	}

	return a
}

func (a *zapEntryAssert) setMessage(message any) *zapEntryAssert {
	if a.message != nil {
		if a.level != nil {
			a.z.reportError(fmt.Sprintf("Msg() can't be called after %s()",
				cases.Title(language.English, cases.NoLower).String(a.level.String())))
		}

		a.z.reportError("Msg() can't be called multiple times")
	}

	a.message = message

	return a
}

func (a *zapEntryAssert) doAssert(ctxMatcher match.Matcher) *Zap {
	if a.done {
		a.z.reportError("Only one of these methods can be called (once): " + zapEntryAssertTerminatingMethods)
	}

	a.done = true
	a.z.pending = nil

	if a.z.orderless {
		observed := a.z.ObservedLogs().All()
		iterateAvailableEntries := func(fn func(i int, entry observer.LoggedEntry) *Zap) *Zap {
			isUsed := func(i int) bool {
				for _, j := range a.z.usedIndexes {
					if i == j {
						return true
					}
				}

				return false
			}

			for i, entry := range observed {
				if isUsed(i) {
					continue
				}

				if z := fn(i, entry); z != nil {
					return z
				}
			}

			return nil
		}

		z := iterateAvailableEntries(func(i int, entry observer.LoggedEntry) *Zap {
			if ok, _ := a.checkEntry(entry, a.name, a.level, a.message, ctxMatcher); ok {
				a.z.usedIndexes = append(a.z.usedIndexes, i)

				return a.z
			}

			return nil
		})
		if z != nil {
			return z
		}

		builder := &strings.Builder{}
		builder.WriteString("No entries matched and there are no more logged entries. Available:\n")
		iterateAvailableEntries(func(i int, entry observer.LoggedEntry) *Zap {
			builder.WriteByte('\t')
			z.formatEntry(builder, entry)
			builder.WriteByte('\n')

			return nil
		})

		a.z.reportError(builder.String())
	}

	if a.z.index == a.z.observedLogs.Len() {
		a.z.reportError("There are no more logged entries")
	}

	entry := a.z.observedLogs.All()[a.z.index]
	a.z.index++

	if ok, desc := a.checkEntry(entry, a.name, a.level, a.message, ctxMatcher); !ok {
		a.z.reportError(strings.TrimPrefix(desc, "Fn(): "))
	}

	return a.z
}

func (a *zapEntryAssert) checkEntry(
	entry observer.LoggedEntry,
	name *string,
	level *zapcore.Level,
	message any,
	ctxMatcher match.Matcher,
) (bool, string) {
	if name != nil {
		if ok, desc := match.Equal(*name).Matches(entry.LoggerName); !ok {
			return ok, a.appendEntry(desc, entry)
		}
	}

	if level != nil {
		if ok, desc := match.Equal(level).Matches(entry.Level); !ok {
			return ok, a.appendEntry(desc, entry)
		}
	}

	if message != nil {
		if ok, desc := match.ToMatcher(message).Matches(entry.Message); !ok {
			return ok, a.appendEntry(desc, entry)
		}
	}

	if ctxMatcher != nil {
		if ok, desc := ctxMatcher.Matches(entry.ContextMap()); !ok {
			desc += "\n"

			return ok, a.appendEntry(desc, entry)
		}
	}

	return true, ""
}

func (a *zapEntryAssert) appendEntry(desc string, entry observer.LoggedEntry) string {
	builder := &strings.Builder{}

	builder.WriteString(desc)
	builder.WriteString("\n\n\tEntry: ")
	a.z.formatEntry(builder, entry)

	return builder.String()
}
