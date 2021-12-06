package qtest_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/borealissw/qtest"
	"github.com/borealissw/qtest/qmock"
	"github.com/stretchr/testify/assert"
)

func TestEquality(t *testing.T) {
	aa := assert.New(t)
	aa.Equal(1, 2)

	expected := qtest.TestResults{}
	var a []int = []int{1, 2, 3}

	for i := range a {
		fmt.Print(i)
	}

	mock := qmock.NewMocker(t)
	defer mock.MockerPanicHandler()

	m := map[interface{}]string{
		mock.Fatal:  "Fatal",
		mock.Fatalf: "Fatalf",
	}

	fmt.Printf("%T - %s", mock.Fatal, m[mock.Fatal])
	mock.Logf("Format %s", "string")
	if mock.LogfCalls.CallCount() != 1 {
		t.Fatalf("Expected 1 call to Logf, found %d", mock.LogfCalls.CallCount())
	}

	call := mock.LogfCalls.Call(0)
	if call.ArgCount() != 2 {
		t.Fatalf("Expected 2 arguments, found %d", call.ArgCount())
	}

	if err := call.VerifyArg(0, "Format %s"); err != nil {
		t.Fatal(err.Error())
	}

	expectedArgs := qmock.NewArgs("Format %s")
	if err := call.VerifyArgs(expectedArgs); err != nil {
		t.Fatal(err.Error())
	}

	mock.Error("Hello")
	if err := mock.ErrorCalls.Call(0).VerifyArg(0, "Hello2"); err != nil {
		t.Fatal(err.Error())
	}

	found := qtest.TestResults{}
	found.Stdout = bytes.NewBufferString("hello")
	fmt.Println("Get's here")
	qtest.CheckResults(&expected, &found, mock)
}
