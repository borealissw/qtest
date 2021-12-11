package qmock

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

const mockerPanic = "TBMockerInducedPanic-e4810650-829f-42c1-9561-f4141434bf35"

func IsMockerPanic(recovery interface{}) bool {
	return recovery == mockerPanic
}

type Recorder struct {
	calls []Call
	lock  sync.Mutex
}

func (recorder *Recorder) AddCall(args ...interface{}) {
	recorder.lock.Lock()
	defer recorder.lock.Unlock()

	recorder.calls = append(recorder.calls, NewCall(args...))
}

func (recorder *Recorder) CallCount() int {
	recorder.lock.Lock()
	defer recorder.lock.Unlock()

	return len(recorder.calls)
}

func (recorder *Recorder) Call(index int) *Call {
	recorder.lock.Lock()
	defer recorder.lock.Unlock()

	return &(recorder.calls[index])
}

func (recorder *Recorder) Reset() {
	recorder.lock.Lock()
	defer recorder.lock.Unlock()

	recorder.calls = nil
}

type Call struct {
	Timestamp time.Time
	Args      []interface{}
}

func NewCall(args ...interface{}) Call {
	return Call{
		Timestamp: time.Now(),
		Args:      args,
	}
}

func (call *Call) ArgCount() int {
	return len(call.Args)
}

func (call *Call) VerifyArg(argIndex int, expected interface{}) error {
	if argIndex >= len(call.Args) {
		return fmt.Errorf("unknown arg: index %v", argIndex)
	}

	actual := call.Args[argIndex]

	if actual == nil && expected == nil {
		return nil
	}

	if (actual != nil) != (expected != nil) {
		if actual != nil {
			return fmt.Errorf(
				"arg %v: expected %T nil, actual %T non-nil",
				argIndex,
				actual,
				actual)
		} else {
			return fmt.Errorf(
				"arg %v: expected %T non-nil, actual %T nil",
				argIndex,
				expected,
				expected)
		}
	}

	if reflect.TypeOf(expected) != reflect.TypeOf(actual) {
		return fmt.Errorf(
			"arg %v: expected type %T, actual type %T",
			argIndex,
			expected,
			call.Args[argIndex])
	}

	if reflect.TypeOf(expected).Comparable() &&
		reflect.TypeOf(actual).Comparable() {
		if actual != expected {
			return fmt.Errorf(
				"arg %v: expected %T '%+v', actual %T '%+v'",
				argIndex,
				expected,
				expected,
				call.Args[argIndex],
				call.Args[argIndex])
		}
	}

	return nil
}

func (call *Call) VerifyArgs(expected ...interface{}) error {
	if len(expected) != len(call.Args) {
		return fmt.Errorf(
			"different arg counts: expected %d, actual %d",
			len(expected),
			len(call.Args))
	}

	for i := range expected {
		if err := call.VerifyArg(i, expected[i]); err != nil {
			return err
		}
	}

	return nil
}

func NewArgs(args ...interface{}) []interface{} {
	return args
}

type TBMocker struct {
	testing.TB
	CleanupCalls Recorder
	ErrorCalls   Recorder
	ErrorfCalls  Recorder
	FailCalls    Recorder
	FailNowCalls Recorder
	FailedCalls  Recorder
	FatalCalls   Recorder
	FatalfCalls  Recorder
	HelperCalls  Recorder
	LogCalls     Recorder
	LogfCalls    Recorder
	NameCalls    Recorder
	SkipCalls    Recorder
	SkipNowCalls Recorder
	SkipfCalls   Recorder
	SkippedCalls Recorder
	TempDirCalls Recorder
	t            testing.TB
	failed       bool
	skipped      bool
}

func NewMocker(t testing.TB) *TBMocker {
	return &TBMocker{
		t:       t,
		failed:  false,
		skipped: false,
	}
}

func (mock *TBMocker) ResetAll() {
	mock.CleanupCalls.Reset()
	mock.ErrorCalls.Reset()
	mock.ErrorfCalls.Reset()
	mock.FailCalls.Reset()
	mock.FailNowCalls.Reset()
	mock.FailedCalls.Reset()
	mock.FatalCalls.Reset()
	mock.FatalfCalls.Reset()
	mock.HelperCalls.Reset()
	mock.LogCalls.Reset()
	mock.LogfCalls.Reset()
	mock.NameCalls.Reset()
	mock.SkipCalls.Reset()
	mock.SkipNowCalls.Reset()
	mock.SkipfCalls.Reset()
	mock.SkippedCalls.Reset()
	mock.TempDirCalls.Reset()

	mock.skipped = false
	mock.failed = false
}

func (mock *TBMocker) MockerPanicHandler() {
	if r := recover(); r != nil {
		if !IsMockerPanic(r) {
			panic(fmt.Sprintf("Unexpected panic: %+v", r))
		}
	}
}

func (mock *TBMocker) Cleanup(cleaner func()) {
	mock.CleanupCalls.AddCall(cleaner)

	mock.t.Cleanup(cleaner)
}

func (mock *TBMocker) Error(args ...interface{}) {
	mock.ErrorCalls.AddCall(args...)
	mock.failed = true
}

func (mock *TBMocker) Errorf(format string, args ...interface{}) {
	args = append([]interface{}{format}, args)
	mock.ErrorfCalls.AddCall(args...)
	mock.failed = true
}

func (mock *TBMocker) Fail() {
	mock.FailCalls.AddCall()
	mock.failed = true
}

func (mock *TBMocker) FailNow() {
	mock.FailNowCalls.AddCall()
	mock.failed = true

	panic(mockerPanic)
}

func (mock *TBMocker) Failed() bool {
	mock.FailedCalls.AddCall()
	return mock.failed
}

func (mock *TBMocker) Fatal(args ...interface{}) {
	mock.FatalCalls.AddCall(args...)
	mock.failed = true

	panic(mockerPanic)
}

func (mock *TBMocker) Fatalf(format string, args ...interface{}) {
	args = append([]interface{}{format}, args)
	mock.FatalfCalls.AddCall(args...)

	mock.failed = true

	panic(mockerPanic)
}

func (mock *TBMocker) Helper() {
	mock.HelperCalls.AddCall()
}

func (mock *TBMocker) Log(args ...interface{}) {
	mock.LogCalls.AddCall(args...)
}

func (mock *TBMocker) Logf(format string, args ...interface{}) {
	args = append([]interface{}{format}, args)
	mock.LogfCalls.AddCall(args...)
}

func (mock *TBMocker) Name() string {
	mock.NameCalls.AddCall()
	return mock.t.Name()
}

func (mock *TBMocker) Skip(args ...interface{}) {
	mock.SkipCalls.AddCall()
	mock.skipped = true
}

func (mock *TBMocker) SkipNow() {
	mock.SkipNowCalls.AddCall()
	mock.skipped = true
}

func (mock *TBMocker) Skipf(format string, args ...interface{}) {
	args = append([]interface{}{format}, args)
	mock.SkipfCalls.AddCall(args...)

	mock.skipped = true
}

func (mock *TBMocker) Skipped() bool {
	mock.SkippedCalls.AddCall()
	return mock.skipped
}

func (mock *TBMocker) TempDir() string {
	mock.TempDirCalls.AddCall()
	return mock.t.TempDir()
}
