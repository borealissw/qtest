package qmock_test

import (
	"fmt"
	"reflect"
	"sync"
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
			assert.Equal(1, len(call.Args))

			if reflect.TypeOf(v).Comparable() {
				assert.Equal(v, call.Args[0])
			}
		})
	}
}

func Test_Recorder_Should_BeAbleToRecordMultipleParametersOfSameType(t *testing.T) {
	for _, v := range testArgs {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			assert := require.New(t)
			r := qmock.Recorder{}

			r.AddCall(v, v, v)
			assert.Equal(1, r.CallCount())

			call := r.Call(0)
			assert.Equal(3, call.ArgCount())

			if reflect.TypeOf(v).Comparable() {
				for i := 0; i < 3; i++ {
					assert.Equal(v, call.Args[i], "Arg %d", i)
				}
			}
		})
	}
}

func Test_Recorder_Should_BeAbleToRecordMultipleParametersOfMixedType(t *testing.T) {
	assert := require.New(t)
	r := qmock.Recorder{}

	r.AddCall(comparable...)
	assert.Equal(1, r.CallCount())

	call := r.Call(0)
	assert.Equal(len(comparable), call.ArgCount())

	for i, v := range testArgs {
		if reflect.TypeOf(v).Comparable() {
			assert.Equal(v, call.Args[i], "Arg %d", i)
		}
	}
}

func Test_Recorder_Should_CanResetCountAfterRecording(t *testing.T) {
	assert := require.New(t)
	r := &qmock.Recorder{}

	r.AddCall(1, 2, 3)
	assert.Equal(1, r.CallCount())

	r.Reset()
	assert.Equal(0, r.CallCount())

	r.AddCall(1, 2, 3)
	assert.Equal(1, r.CallCount())
}

func Test_Call_VerifyArg_Should_ReturnNoErrorIfBothValuesAreNil(t *testing.T) {
	assert := require.New(t)
	call := qmock.NewCall(nil)

	err := call.VerifyArg(0, nil)

	assert.NoError(err)
}

func Test_Call_VerifyArg_Should_ReturnNoErrorIfComparableValuesAreEqual(t *testing.T) {
	for _, v := range comparable {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			assert := require.New(t)
			call := qmock.NewCall(v)

			err := call.VerifyArg(0, v)

			assert.NoError(err)
		})
	}
}

func Test_Call_VerifyArg_Should_ReturnNoErrorIfIncomparableTypesAreEqualAndNilIsSame(t *testing.T) {
	for _, v := range incomparable {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			assert := require.New(t)
			call := qmock.NewCall(v)

			err := call.VerifyArg(0, v)

			assert.NoError(err)
		})
	}
}

func Test_Call_VerifyArg_Should_ReturnErrorIfVerifyingOutOfRangeArgument(t *testing.T) {
	assert := require.New(t)
	call := qmock.NewCall()

	err := call.VerifyArg(0, 1)

	assert.EqualError(err, "unknown arg: index 0")
}

func Test_Call_VerifyArg_Should_ReturnErrorIfVerifyingArgumentOfWrongType(t *testing.T) {
	assert := require.New(t)
	call := qmock.NewCall(0)

	err := call.VerifyArg(0, "string")

	assert.EqualError(err, "arg 0: expected type string, actual type int")
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
			assert := require.New(t)

			call := qmock.NewCall(v)
			test := testValues[i]
			err := call.VerifyArg(0, test)

			expected := fmt.Sprintf(
				"arg 0: expected %T '%+v', actual %T '%+v'",
				test, test,
				v, v)
			assert.EqualError(err, expected)
		})
	}
}

func Test_Call_VerifyArg_Should_ReturnErrorIfIncomparableValuesDiffer(t *testing.T) {
	for _, v := range incomparable {
		t.Run(fmt.Sprintf("%T/nil", v), func(t *testing.T) {
			assert := require.New(t)
			call := qmock.NewCall(v)

			err := call.VerifyArg(0, nil)

			expected := fmt.Sprintf("arg 0: expected %T nil, actual %T non-nil", v, v)
			assert.EqualError(err, expected)
		})
	}

	for _, v := range incomparable {
		t.Run(fmt.Sprintf("%T/non-nil", v), func(t *testing.T) {
			assert := require.New(t)
			call := qmock.NewCall(nil)

			err := call.VerifyArg(0, v)

			expected := fmt.Sprintf("arg 0: expected %T non-nil, actual %T nil", v, v)
			assert.EqualError(err, expected)
		})
	}
}

func Test_Call_VerifyArgs_ShouldReturnNoErrorIfAllArgumentsAreValid(t *testing.T) {
	assert := require.New(t)
	call := qmock.NewCall(testArgs...)

	err := call.VerifyArgs(testArgs...)
	assert.NoError(err)
}

func Test_Call_VerifyArgs_ShouldReturnErrorIfArgumentLengthsDiffer(t *testing.T) {
	assert := require.New(t)
	call := qmock.NewCall(1, 2, 3)

	err := call.VerifyArgs(1, 3)

	expected := "different arg counts: expected 2, actual 3"
	assert.EqualError(err, expected)
}

func Test_Call_VerifyArgs_ShouldReturnErrorIfAnyArgumentDiffers(t *testing.T) {
	for i, v := range testArgs {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			assert := require.New(t)

			testArgsCopy := make([]interface{}, len(testArgs))
			copy(testArgsCopy, testArgs)
			testArgsCopy[i] = nil

			call := qmock.NewCall(testArgs...)
			err := call.VerifyArgs(testArgsCopy...)

			expected := fmt.Sprintf("arg %d: expected %T nil, actual %T non-nil", i, v, v)
			assert.EqualError(err, expected)
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
			assert := require.New(t)

			callCount := test()

			assert.Equal(1, callCount)
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
			assert := require.New(t)

			panicTriggered := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicTriggered = true
						assert.True(qmock.IsMockerPanic(r))
					}
				}()

				test.action()
				assert.FailNow("Expected action didn't panic")
			}()

			assert.True(panicTriggered, "Invalid test - panic not triggered")

			callCount := test.recorder.CallCount()
			assert.Equal(1, callCount)
		})
	}
}

func Test_Mocker_ShouldHandleDataRacesCorrectlyWhenRecordingCalls(t *testing.T) {
	var wg sync.WaitGroup
	mocker := qmock.NewMocker(t)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(callNo int) {
			mocker.Logf("Call %d", callNo)
			wg.Done()
		}(i)
	}

	wg.Wait()
	require.Equal(t, 1000, mocker.LogfCalls.CallCount())
}

func Test_Mocker_ShouldHandleDataRacesCorrectlyWhenRecordingAndAccessingCalls(t *testing.T) {
	var wg sync.WaitGroup
	mocker := qmock.NewMocker(t)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(callNo int) {
			mocker.Logf("Call %d: count now %d", callNo, mocker.LogfCalls.CallCount())
			wg.Done()
		}(i)
	}

	wg.Wait()
	require.Equal(t, 1000, mocker.LogfCalls.CallCount())
}

func Test_PanicHandling_RecoversFromTestTerminatingMethodCall(t *testing.T) {
	assert := require.New(t)
	mocker := qmock.NewMocker(t)

	func() {
		defer mocker.MockerPanicHandler()

		mocker.Fatal("Mocker panic handler test")
		assert.FailNow("Expected action didn't panic")
	}()

	callCount := mocker.FatalCalls.CallCount()
	assert.Equal(1, callCount)
}

func Test_PanicHandling_TriggersPanifIfPanicNotFromMocker(t *testing.T) {
	assert := require.New(t)
	mocker := qmock.NewMocker(t)

	panicTriggered := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				assert.False(qmock.IsMockerPanic(r))
				assert.Equal("Unexpected panic: Not from TBMocker", r)

				panicTriggered = true
			}
		}()

		func() {
			defer mocker.MockerPanicHandler()

			panic("Not from TBMocker")
		}()
	}()

	assert.True(panicTriggered, "Invalid test - panic not triggered")
}

func Test_Mocker_ResetAll_Should_ResetEveryRecorder(t *testing.T) {
	assert := require.New(t)
	mocker := qmock.NewMocker(t)

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
		assert.FailNow("Expected action didn't panic")
	}()

	mocker.Skip("Mocker panic handler test")

	mocker.ResetAll()

	for name, recorder := range recorders {
		assert.Zero(recorder.CallCount(), "Recorder %s", name)
	}

	assert.False(mocker.Failed(), "Failed flag not cleared")
	assert.False(mocker.Skipped(), "Skipped flag not cleared")
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
			assert := require.New(t)
			mocker := qmock.NewMocker(t)

			panicTriggered := false

			func() {
				defer func() {
					if r := recover(); r != nil {
						if !qmock.IsMockerPanic(r) {
							t.Fatalf("Non-mocker panic: %+v", r)
						}
					}

					panicTriggered = true
				}()

				test(mocker)
				assert.FailNow("Expected action didn't panic")
			}()

			assert.True(panicTriggered, "Invalid test - panic not triggered")
			assert.True(mocker.Failed(), "Method %s not registered as failure", name)
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
			assert := require.New(t)
			mocker := qmock.NewMocker(t)
			test(mocker)

			assert.True(mocker.Skipped(), "Method %s not registering test as skipped", name)
		})
	}
}

func Test_SideEffects_CleanupMethodsShouldBeCalled(t *testing.T) {
	assert := require.New(t)

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

	assert.True(cleaner1Called, "Cleaner method 1 not called")
	assert.True(cleaner2Called, "Cleaner method 2 not called")
}
