package qmock_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/borealissw/qtest/qmock"
)

type testStructure struct {
	i int
	s string
}

type any = interface{}

var testInt int = 1
var testDouble float64 = 1.234
var testString string = "string"
var testStruct testStructure = testStructure{i: 1, s: "s"}
var testPointer *testStructure = &testStruct
var testFunction func() = func() {}
var testSlice []any = []any{1, "two", 3.0}

var comparable []any = []any{
	testInt,
	testDouble,
	testString,
	testStruct,
	testPointer}

var incomparable []any = []any{
	testFunction,
	testSlice}

var testArgs = append(comparable, incomparable...)

func Test_Recorder_Should_BeAbleToRecordAnyParameter(t *testing.T) {
	for _, v := range testArgs {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			assert := require.New(t)
			r := qmock.Recorder{}
			r.AddCall(v)

			assert.Equal(1, r.CallCount())

			call := r.Call(0)
			if len(call.Args) != 1 {
				t.Fatalf("Expected args count 1, actual %d", len(call.Args))
			}

			if reflect.TypeOf(v).Comparable() {
				if call.Args[0] != v {
					t.Fatalf(
						"Expected %T '%v', actual %T '%v'",
						call.Args[0], call.Args[0],
						v, v)
				}
			}
		})
	}
}

func Test_Recorder_Should_BeAbleToRecordMultipleParametersOfSameType(t *testing.T) {
	for _, v := range testArgs {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			r := qmock.Recorder{}
			r.AddCall(v, v, v)

			if r.CallCount() != 1 {
				t.Fatalf("Expected call count 1, actual %d", r.CallCount())
			}

			call := r.Call(0)
			if len(call.Args) != 3 {
				t.Fatalf("Expected args count 3, actual %d", len(call.Args))
			}

			if reflect.TypeOf(v).Comparable() {
				for i := 0; i < 3; i++ {
					if call.Args[i] != v {
						t.Fatalf(
							"Expected arg %d %T '%v', actual %T '%v'",
							i,
							call.Args[0], call.Args[0],
							v, v)
					}
				}
			}
		})
	}
}

func Test_Recorder_Should_BeAbleToRecordMultipleParametersOfMixedType(t *testing.T) {
	r := qmock.Recorder{}
	r.AddCall(comparable...)

	if r.CallCount() != 1 {
		t.Fatalf("Expected call count 1, actual %d", r.CallCount())
	}

	call := r.Call(0)
	if len(call.Args) != len(comparable) {
		t.Fatalf("Expected args count %d, actual %d", len(comparable), len(call.Args))
	}

	for i, v := range testArgs {
		if reflect.TypeOf(v).Comparable() {
			if call.Args[i] != v {
				t.Fatalf(
					"Expected arg %d %T '%v', actual %T '%v'",
					i,
					call.Args[0], call.Args[0],
					v, v)
			}
		}
	}
}

func Test_Recorder_Should_CanResetCountAfterRecording(t *testing.T) {
	r := &qmock.Recorder{}

	verify := func(recorder *qmock.Recorder, expected int) {
		callCount := recorder.CallCount()
		if callCount != expected {
			t.Fatalf("Expected call count %d, actual %d", expected, callCount)
		}
	}

	r.AddCall(1, 2, 3)
	verify(r, 1)

	r.Reset()
	verify(r, 0)

	r.AddCall(1, 2, 3)
	verify(r, 1)
}

func Test_Call_VerifyArg_Should_ReturnNoErrorIfBothValuesAreNil(t *testing.T) {
	call := qmock.NewCall(nil)
	err := call.VerifyArg(0, nil)
	if err != nil {
		t.Fatalf("Expected no error, actual %+v", err)
	}
}

func Test_Call_VerifyArg_Should_ReturnNoErrorIfComparableValuesAreEqual(t *testing.T) {
	for _, v := range comparable {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			call := qmock.NewCall(v)

			err := call.VerifyArg(0, v)
			if err != nil {
				t.Fatalf("Expected no error, actual %+v", err)
			}
		})
	}
}

func Test_Call_VerifyArg_Should_ReturnNoErrorIfIncomparableTypesAreEqualAndNilIsSame(t *testing.T) {
	for _, v := range incomparable {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			call := qmock.NewCall(v)

			err := call.VerifyArg(0, v)
			if err != nil {
				t.Fatalf("Expected no error, actual %+v", err)
			}
		})
	}
}

func Test_Call_VerifyArg_Should_ReturnErrorIfVerifyingOutOfRangeArgument(t *testing.T) {
	call := qmock.NewCall()

	err := call.VerifyArg(0, 1)
	if err == nil {
		t.Fatal("Expected error, actual nil")
	}

	if err.Error() != "unknown arg: index 0" {
		t.Fatalf("Expected 'unknown arg: index 0', actual '%s'", err.Error())
	}
}

func Test_Call_VerifyArg_Should_ReturnErrorIfVerifyingArgumentOfWrongType(t *testing.T) {
	call := qmock.NewCall(0)

	err := call.VerifyArg(0, "string")
	if err == nil {
		t.Fatal("Expected error, actual nil")
	}

	expected := "arg 0: expected type string, actual type int"
	if err.Error() != expected {
		t.Fatalf("Expected '%s', actual '%s'", expected, err.Error())
	}
}

func Test_Call_VerifyArg_Should_ReturnErrorIfComparableValuesDiffer(t *testing.T) {
	newStruct := testStructure{i: 22, s: "new"}
	testValues := []any{
		11,
		22.0,
		"different string",
		newStruct,
		&newStruct,
	}

	for i, v := range comparable {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			call := qmock.NewCall(v)

			test := testValues[i]
			err := call.VerifyArg(0, test)
			if err == nil {
				t.Fatal("Expected error, actual nil")
			}

			expected := fmt.Sprintf(
				"arg 0: expected %T '%+v', actual %T '%+v'",
				test, test,
				v, v)
			if err.Error() != expected {
				t.Fatalf("Expected '%s', actual '%s'", expected, err.Error())
			}
		})
	}
}

func Test_Call_VerifyArg_Should_ReturnErrorIfIncomparableValuesDiffer(t *testing.T) {
	for _, v := range incomparable {
		t.Run(fmt.Sprintf("%T/nil", v), func(t *testing.T) {
			call := qmock.NewCall(v)
			err := call.VerifyArg(0, nil)
			if err == nil {
				t.Fatal("Expected error, actual nil")
			}

			expected := fmt.Sprintf("arg 0: expected %T nil, actual %T non-nil", v, v)
			if err.Error() != expected {
				t.Fatalf("Expected '%s', actual '%s'", expected, err.Error())
			}
		})
	}

	for _, v := range incomparable {
		t.Run(fmt.Sprintf("%T/non-nil", v), func(t *testing.T) {
			call := qmock.NewCall(nil)

			err := call.VerifyArg(0, v)
			if err == nil {
				t.Fatal("Expected error, actual nil")
			}

			expected := fmt.Sprintf("arg 0: expected %T non-nil, actual %T nil", v, v)
			if err.Error() != expected {
				t.Fatalf("Expected '%s', actual '%s'", expected, err.Error())
			}
		})
	}
}

func Test_Call_VerifyArgs_ShouldReturnNoErrorIfAllArgumentsAreValid(t *testing.T) {
	call := qmock.NewCall(testArgs...)

	err := call.VerifyArgs(testArgs...)
	if err != nil {
		t.Fatalf("Expected error, actual %+v", err)
	}
}

func Test_Call_VerifyArgs_ShouldReturnErrorIfArgumentLengthsDiffer(t *testing.T) {
	call := qmock.NewCall(1, 2, 3)

	err := call.VerifyArgs(1, 3)
	if err == nil {
		t.Fatal("Expected error, actual nil")
	}

	expected := "different arg counts: expected 2, actual 3"
	if err.Error() != expected {
		t.Fatalf("Expected '%s', actual '%s'", expected, err.Error())
	}
}

func Test_Call_VerifyArgs_ShouldReturnErrorIfAnyArgumentDiffers(t *testing.T) {
	for i, v := range testArgs {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			testArgsCopy := make([]interface{}, len(testArgs))
			copy(testArgsCopy, testArgs)
			testArgsCopy[i] = nil

			call := qmock.NewCall(testArgs...)
			err := call.VerifyArgs(testArgsCopy...)
			if err == nil {
				t.Fatal("Expected error, actual nil")
			}

			expected := fmt.Sprintf("arg %d: expected %T nil, actual %T non-nil", i, v, v)
			if err.Error() != expected {
				t.Fatalf("Expected '%s', actual '%s'", expected, err.Error())
			}
		})
	}
}

func Test_Mocker_AllReturningMethodsShouldRecordACallWithoutFailure(t *testing.T) {
	mocker := qmock.NewMocker(t)

	tests := map[string]func() int{
		"Cleanup": func() int {
			mocker.Cleanup(func() {})
			return mocker.CleanupCalls.CallCount()
		},
		"Error": func() int {
			mocker.Error(1, 2, 3)
			return mocker.ErrorCalls.CallCount()
		},
		"Errorf": func() int {
			mocker.Errorf("%d - %d - %d", 1, 2, 3)
			return mocker.ErrorfCalls.CallCount()
		},
		"Fail": func() int {
			mocker.Fail()
			return mocker.FailCalls.CallCount()
		},
		"Failed": func() int {
			mocker.Failed()
			return mocker.FailedCalls.CallCount()
		},
		"HelperCalls": func() int {
			mocker.Helper()
			return mocker.HelperCalls.CallCount()
		},
		"LogCalls": func() int {
			mocker.Log(1, 2, 3)
			return mocker.LogCalls.CallCount()
		},
		"LogfCalls": func() int {
			mocker.Logf("%d - %d - %d", 1, 2, 3)
			return mocker.LogfCalls.CallCount()
		},
		"NameCalls": func() int {
			mocker.Name()
			return mocker.NameCalls.CallCount()
		},
		"SkipCalls": func() int {
			mocker.Skip(1, 2, 3)
			return mocker.SkipCalls.CallCount()
		},
		"SkipNowCalls": func() int {
			mocker.SkipNow()
			return mocker.SkipNowCalls.CallCount()
		},
		"SkipfCalls": func() int {
			mocker.Skipf("%d - %d - %d", 1, 2, 3)
			return mocker.SkipfCalls.CallCount()
		},
		"SkippedCalls": func() int {
			mocker.Skipped()
			return mocker.SkippedCalls.CallCount()
		},
		"TempDirCalls": func() int {
			mocker.TempDir()
			return mocker.TempDirCalls.CallCount()
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			callCount := test()

			if callCount != 1 {
				t.Fatalf("Expected 1 call, actual %d calls", callCount)
			}
		})
	}
}

func Test_Mocker_AllTerminatingMethodsShouldRecordACallAndFail(t *testing.T) {
	mocker := qmock.NewMocker(t)

	tests := map[string]struct {
		action   func()
		recorder *qmock.Recorder
	}{
		"FailNow": {
			action: func() {
				mocker.FailNow()
			},
			recorder: &mocker.FailNowCalls,
		},
		"Fatal": {
			action: func() {
				mocker.Fatal(1, 2, 3)
			},
			recorder: &mocker.FatalCalls,
		},
		"Fatalf": {
			action: func() {
				mocker.Fatalf("%d - %d - %d", 1, 2, 3)
			},
			recorder: &mocker.FatalfCalls,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			func() {
				defer func() {
					if r := recover(); r != nil {
						if !qmock.IsMockerPanic(r) {
							t.Fatalf("Non-mocker panic: %+v", r)
						}
					}
				}()

				test.action()
				t.Fatal("Expected action didn't panic")
			}()

			callCount := test.recorder.CallCount()
			if callCount != 1 {
				t.Fatalf("Expected 1 call, actual %d calls", callCount)
			}
		})
	}
}

func Test_PanicHandling_RecoversFromTestTerminatingMethodCall(t *testing.T) {
	mocker := qmock.NewMocker(t)

	func() {
		defer mocker.MockerPanicHandler()

		mocker.Fatal("Mocker panic handler test")
		t.Fatal("Expected action didn't panic")
	}()

	callCount := mocker.FatalCalls.CallCount()
	if callCount != 1 {
		t.Fatalf("Test invalid: wxpected 1 call, actual %d calls", callCount)
	}
}

func Test_PanicHandling_TriggersPanifIfPanicNotFromMocker(t *testing.T) {
	mocker := qmock.NewMocker(t)

	panicTriggered := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				if qmock.IsMockerPanic(r) {
					t.Fatalf("Unexpected mocker panic: %+v", r)
				}

				expected := "Unexpected panic: Not from TBMocker"
				if r != expected {
					t.Fatalf("Expected panic '%s', actual %+v", expected, r)
				}

				panicTriggered = true
			}
		}()

		func() {
			defer mocker.MockerPanicHandler()

			panic("Not from TBMocker")
		}()
	}()

	if !panicTriggered {
		t.Fatal("Expected panic not encountered")
	}
}

func Test_Mocker_ResetAll_Should_ResetEveryRecorder(t *testing.T) {
	mocker := qmock.NewMocker(t)

	verifyReset := func(name string, recorder *qmock.Recorder) {
		callCount := recorder.CallCount()
		if callCount != 0 {
			t.Fatalf("%s: expected call count 0, actual %d", name, callCount)
		}
	}

	recorders := map[string]*qmock.Recorder{
		"Cleanup": &mocker.CleanupCalls,
		"Error":   &mocker.ErrorCalls,
		"Errorf":  &mocker.ErrorfCalls,
		"Fail":    &mocker.FailCalls,
		"FailNow": &mocker.FailNowCalls,
		"Failed":  &mocker.FailedCalls,
		"Fatal":   &mocker.FatalCalls,
		"Fatalf":  &mocker.FatalfCalls,
		"Helper":  &mocker.HelperCalls,
		"Log":     &mocker.LogCalls,
		"Logf":    &mocker.LogfCalls,
		"Name":    &mocker.NameCalls,
		"Skip":    &mocker.SkipCalls,
		"SkipNow": &mocker.SkipNowCalls,
		"Skipf":   &mocker.SkipfCalls,
		"Skipped": &mocker.SkippedCalls,
		"TempDir": &mocker.TempDirCalls,
	}

	for _, recorder := range recorders {
		recorder.AddCall(1, 2, 3)
	}

	func() {
		defer mocker.MockerPanicHandler()

		mocker.Fatal("Mocker panic handler test")
		t.Fatal("Expected action didn't panic")
	}()

	mocker.Skip("Mocker panic handler test")

	mocker.ResetAll()

	for name, recorder := range recorders {
		verifyReset(name, recorder)
	}

	if mocker.Failed() {
		t.Fatal("Failed flag not cleared")
	}

	if mocker.Skipped() {
		t.Fatal("Skipped flag not cleared")
	}
}

func Test_SideEffects_VerifyFailedBehaviourAfterFailingTrigger(t *testing.T) {
	tests := map[string]func(*qmock.TBMocker){
		"FailNow": func(mocker *qmock.TBMocker) {
			mocker.FailNow()
		},
		"Fatal": func(mocker *qmock.TBMocker) {
			mocker.Fatal(1, 2, 3)
		},
		"Fatalf": func(mocker *qmock.TBMocker) {
			mocker.Fatalf("%d - %d - %d", 1, 2, 3)
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mocker := qmock.NewMocker(t)
			func() {
				defer func() {
					if r := recover(); r != nil {
						if !qmock.IsMockerPanic(r) {
							t.Fatalf("Non-mocker panic: %+v", r)
						}
					}
				}()

				test(mocker)
				t.Fatal("Expected action didn't panic")
			}()

			if !mocker.Failed() {
				t.Fatalf("Mocker does not register test has failed")
			}
		})
	}
}

func Test_SideEffects_VerifySkippedBehaviourAfterSkippingTrigger(t *testing.T) {
	tests := map[string]func(*qmock.TBMocker){
		"Skip": func(mocker *qmock.TBMocker) {
			mocker.Skip(1, 2, 3)
		},
		"SkipNow": func(mocker *qmock.TBMocker) {
			mocker.SkipNow()
		},
		"Skipf": func(mocker *qmock.TBMocker) {
			mocker.Skipf("%d - %d - %d", 1, 2, 3)
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mocker := qmock.NewMocker(t)
			test(mocker)

			if !mocker.Skipped() {
				t.Fatalf("Mocker does not register test has skipped")
			}
		})
	}
}

func Test_SideEffects_CleanupMethodsShouldBeCalled(t *testing.T) {
	cleaner1Called := false
	cleaner2Called := false

	cleaner := func(flag *bool) {
		*flag = true
	}

	t.Run("runner", func(t *testing.T) {
		mocker := qmock.NewMocker(t)
		mocker.Cleanup(func() {
			cleaner(&cleaner1Called)
		})
		mocker.Cleanup(func() {
			cleaner(&cleaner2Called)
		})
	})

	if !cleaner1Called {
		t.Errorf("Cleaner method 1 didn't get called")
	}

	if !cleaner2Called {
		t.Errorf("Cleaner method 2 didn't get called")
	}
}
