package server

import (
	"reflect"
	"runtime"
	"testing"
)

func assertEquals(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf(
			"Expect equals, but not at %v:%v: Expected: %#v , Actual: %#v",
			file,
			line,
			expected,
			actual,
		)
	}
}

func assertSame(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf(
			"Expect same, but not at %v:%v: Expected: %#v , Actual: %#v",
			file,
			line,
			expected,
			actual,
		)
	}
}

func assertNotSame(t *testing.T, expected, actual interface{}) {
	if expected == actual {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf(
			"Expect not same but is at %v:%v: Expected: %#v , Actual: %#v",
			file,
			line,
			expected,
			actual,
		)
	}
}

func TestExcludingCopyStructTag(t *testing.T) {
	type testStruct struct {
		Field1 int `excludingcopy:"cond1"`
		Field2 int `excludingcopy:"cond2"`
		Field3 int `excludingcopy:"cond1,cond2"`
		Field4 int `anothertag:"cond1"`
		Field5 int `excludingcopy:"cond11"`
		Field6 int
		field7 int
	}

	dst := testStruct{
		Field1: 1,
		Field2: 2,
		Field3: 3,
		Field4: 4,
		Field5: 5,
		Field6: 6,
		field7: 7,
	}
	src := testStruct{
		Field1: 111,
		Field2: 222,
		Field3: 333,
		Field4: 444,
		Field5: 555,
		Field6: 666,
		field7: 777,
	}
	expected := testStruct{
		Field1: 1,
		Field2: 222,
		Field3: 3,
		Field4: 444,
		Field5: 555,
		Field6: 666,
		field7: 7,
	}

	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, "cond1"),
	)
	assertEquals(t, expected, dst)
}

func TestExcludingCopyStructSimpleTypes(t *testing.T) {
	type testStruct struct {
		Field1 int
		Field2 int64
		Field3 float32
		Field4 string
	}

	dst := testStruct{}
	src := testStruct{
		Field1: 12345,
		Field2: 1234567890123456789,
		Field3: 123.4567,
		Field4: "field4",
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
}

func TestExcludingCopyStructTypeDef(t *testing.T) {
	type mytype string
	type testStruct struct {
		Field1 mytype
	}

	dst := testStruct{}
	src := testStruct{
		Field1: mytype("field5"),
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
}

func TestExcludingCopyStructArray(t *testing.T) {
	type testStruct struct {
		Field1 [2]int
	}

	dst := testStruct{}
	src := testStruct{
		Field1: [2]int{1, 2},
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
}

func TestExcludingCopyStructSlice(t *testing.T) {
	type testStruct struct {
		Field1 []int
	}

	dst := testStruct{}
	src := testStruct{
		Field1: []int{3, 4, 5, 6, 7},
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
}

func TestExcludingCopyStructMap(t *testing.T) {
	type testStruct struct {
		Field1 map[string]int
	}

	dst := testStruct{}
	src := testStruct{
		Field1: map[string]int{"key1": 1, "key2": 2},
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
}

func TestExcludingCopyStructInterface(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	dst := testStruct{}
	v := "some string"
	src := testStruct{
		Field1: &v,
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
}

func TestExcludingCopyStructStruct(t *testing.T) {
	type nestStruct struct {
		Value int
	}
	type testStruct struct {
		Field1 nestStruct
	}

	dst := testStruct{}
	src := testStruct{
		Field1: nestStruct{Value: 1234567},
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
}

func TestExcludingCopyStructAnonymousStruct(t *testing.T) {
	type testStruct struct {
		Field1 struct{Value int}
	}

	dst := testStruct{}
	src := testStruct{
		Field1: struct{Value int}{Value: 1234567},
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
}

func TestExcludingCopyStructPointerToSimple(t *testing.T) {
	type testStruct struct {
		Field1 *int
	}

	dst := testStruct{}
	v := 1
	src := testStruct{
		Field1: &v,
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
	assertNotSame(t, src.Field1, dst.Field1)
}

func TestExcludingCopyStructPointerToSimpleOverwrite(t *testing.T) {
	type testStruct struct {
		Field1 *int
	}

	v1 := 1
	v2 := 2
	dst := testStruct{
		Field1: &v1,
	}
	src := testStruct{
		Field1: &v2,
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
	assertSame(t, &v1, dst.Field1)
}

func TestExcludingCopyStructPointerToStruct(t *testing.T) {
	type nestStruct struct {
		Value int
	}
	type testStruct struct {
		Field1 *nestStruct
	}

	dst := testStruct{}
	v := nestStruct{
		Value: 1,
	}
	src := testStruct{
		Field1: &v,
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
	assertNotSame(t, src.Field1, dst.Field1)
}

func TestExcludingCopyStructPointerToStructOverwrite(t *testing.T) {
	type nestStruct struct {
		Value int
	}
	type testStruct struct {
		Field1 *nestStruct
	}

	v1 := nestStruct{
		Value: 1,
	}
	v2 := nestStruct{
		Value: 2,
	}
	dst := testStruct{
		Field1: &v1,
	}
	src := testStruct{
		Field1: &v2,
	}
	assertEquals(
		t,
		nil,
		ExcludingCopy(&dst, &src, ""),
	)
	assertEquals(t, src, dst)
	assertSame(t, &v1, dst.Field1)
}

