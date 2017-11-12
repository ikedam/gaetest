package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func expectEquals(t *testing.T, expected, actual interface{}) {
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

func expectNotEquals(t *testing.T, expected, actual interface{}) {
	if reflect.DeepEqual(expected, actual) {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf(
			"Expect not equals, but does at %v:%v: Expected: %#v , Actual: %#v",
			file,
			line,
			expected,
			actual,
		)
	}
}

func isSame(expected, actual interface{}) bool {
	refExpected := reflect.ValueOf(expected)
	refActual := reflect.ValueOf(actual)

	if refExpected.Kind() != refActual.Kind() {
		return false
	}

	switch refExpected.Kind() {
	case reflect.Slice:
		// Slice consists of the head pointer and the length.
		if refExpected.Len() != refActual.Len() {
			return false
		}
		return refExpected.Pointer() == refActual.Pointer()
	case reflect.Map:
		return refExpected.Pointer() == refActual.Pointer()
	case reflect.Ptr:
		return expected == actual
	}
	// useless comparison (passed with by value)
	return false
}

func identityString(v interface{}) string {
	if v == nil {
		return "nil"
	}

	refV := reflect.ValueOf(v)
	typeV := reflect.TypeOf(v)
	switch typeV.Kind() {
	case reflect.Slice:
		return fmt.Sprintf(
			"Slice[ptr=%p, len=%v, contents=%#v]",
			refV.Pointer(),
			refV.Len(),
			v,
		)
	case reflect.Map:
		return fmt.Sprintf(
			"Map[ptr=%p, contents=%#v]",
			refV.Pointer(),
			v,
		)
	case reflect.Ptr:
		return fmt.Sprintf(
			"Ptr[ptr=%p, contents=%v]",
			refV.Pointer(),
			identityString(refV.Elem().Interface()),
		)
	}
	return fmt.Sprintf("%#v", v)
}

func expectSame(t *testing.T, expected, actual interface{}) {
	if !isSame(expected, actual) {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf(
			"Expect same, but not at %v:%v: Expected: %v , Actual: %v",
			file,
			line,
			identityString(expected),
			identityString(actual),
		)
	}
}

func expectNotSame(t *testing.T, expected, actual interface{}) {
	if isSame(expected, actual) {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf(
			"Expect not same but is at %v:%v: Expected: %v , Actual: %v",
			file,
			line,
			identityString(expected),
			identityString(actual),
		)
	}
}

func expectSliceInSameAddr(t *testing.T, expected, actual interface{}) {
	refExpected := reflect.ValueOf(expected)
	refActual := reflect.ValueOf(actual)

	if refExpected.Kind() != reflect.Slice {
		panic(fmt.Sprintf("expected should be slice, but is %T", expected))
	}
	if refActual.Kind() != reflect.Slice {
		panic(fmt.Sprintf("actual should be slice, but is %T", actual))
	}

	if refExpected.Pointer() != refActual.Pointer() {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf(
			"Expect same addr but differ at %v:%v: Expected: %v , Actual: %v",
			file,
			line,
			identityString(expected),
			identityString(actual),
		)
	}
}

func expectSliceNotInSameAddr(t *testing.T, expected, actual interface{}) {
	refExpected := reflect.ValueOf(expected)
	refActual := reflect.ValueOf(actual)

	if refExpected.Pointer() == refActual.Pointer() {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf(
			"Expect different addr but is same at %v:%v: Expected: %v , Actual: %v",
			file,
			line,
			identityString(expected),
			identityString(actual),
		)
	}
}

func TestIsSameSlice(t *testing.T) {
	v1 := append(make([]int, 0, 3), 1, 2)
	v2 := v1
	expectSame(t, v1, v2)

	v3 := []int{1, 2}

	expectEquals(t, v1, v3)
	expectNotSame(t, v1, v3)

	v4 := append(v1, 3)
	expectNotSame(t, v1, v4)

	v1[0] = 0
	expectSame(t, v1, v2)
}

func TestIsSameMap(t *testing.T) {
	v1 := map[string]int{"key1": 1, "key2": 2}
	v2 := v1
	expectSame(t, v1, v2)

	v3 := map[string]int{"key1": 1, "key2": 2}

	expectEquals(t, v1, v3)
	expectNotSame(t, v1, v3)
}

func TestIsSamePtr(t *testing.T) {
	v1 := 1
	v2 := &v1
	v3 := &v1
	expectSame(t, v2, v3)

	v4 := 1
	v5 := &v4

	expectNotSame(t, v2, v5)
}

func TestProtectingCopyStructTag(t *testing.T) {
	type testStruct struct {
		Field1 int `protectfor:"cond1"`
		Field2 int `protectfor:"cond2"`
		Field3 int `protectfor:"cond1,cond2"`
		Field4 int `anothertag:"cond1"`
		Field5 int `protectfor:"cond11"`
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

	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, "cond1"),
	)
	expectEquals(t, expected, dst)
}

func TestProtectingCopyStructSimpleTypes(t *testing.T) {
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
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructTypeDef(t *testing.T) {
	type mytype string
	type testStruct struct {
		Field1 mytype
	}

	dst := testStruct{}
	src := testStruct{
		Field1: mytype("field5"),
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructArray(t *testing.T) {
	type testStruct struct {
		Field1 [2]int
	}

	dst := testStruct{}
	src := testStruct{
		Field1: [2]int{1, 2},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructSlice(t *testing.T) {
	type testStruct struct {
		Field1 []int
	}

	dst := testStruct{}
	src := testStruct{
		Field1: []int{3, 4, 5, 6, 7},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
}

func TestProtectingCopyStructSliceNil(t *testing.T) {
	type testStruct struct {
		Field1 []int
	}

	dst := testStruct{
		Field1: []int{3, 4, 5, 6, 7},
	}
	src := testStruct{
		Field1: nil,
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructSliceOverwrite(t *testing.T) {
	type testStruct struct {
		Field1 []int
	}

	v := []int{1, 2, 3, 4, 5}
	dst := testStruct{
		Field1: v,
	}
	src := testStruct{
		Field1: []int{3, 4, 5, 6, 7},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectSame(t, v, dst.Field1)
}

func TestProtectingCopyStructSliceReplaceShrink(t *testing.T) {
	type testStruct struct {
		Field1 []int
	}

	v := []int{1, 2, 3, 4, 5, 6}
	dst := testStruct{
		Field1: v,
	}
	src := testStruct{
		Field1: []int{3, 4, 5, 6, 7},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectNotSame(t, v, dst.Field1)
	expectSliceNotInSameAddr(t, v, dst.Field1)
}

func TestProtectingCopyStructSliceReplaceExtend(t *testing.T) {
	type testStruct struct {
		Field1 []int
	}

	v := make([]int, 0, 5)
	dst := testStruct{
		Field1: v,
	}
	src := testStruct{
		Field1: []int{3, 4, 5, 6, 7},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectNotSame(t, v, dst.Field1)
	expectSliceNotInSameAddr(t, v, dst.Field1)
}

func TestProtectingCopyStructMap(t *testing.T) {
	type testStruct struct {
		Field1 map[string]int
	}

	dst := testStruct{}
	src := testStruct{
		Field1: map[string]int{"key1": 1, "key2": 2},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
}

func TestProtectingCopyStructMapNil(t *testing.T) {
	type testStruct struct {
		Field1 map[string]int
	}

	dst := testStruct{
		Field1: map[string]int{"key1": 1, "key2": 2},
	}
	src := testStruct{
		Field1: nil,
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructMapOverwrite(t *testing.T) {
	type testStruct struct {
		Field1 map[string]int
	}

	v1 := map[string]int{"key1": 1, "key2": 103, "key4": 4}
	dst := testStruct{
		Field1: v1,
	}
	src := testStruct{
		Field1: map[string]int{"key1": 1, "key2": 2, "key3": 3},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectSame(t, v1, dst.Field1)
}

func TestProtectingCopyStructInterfaceSimple(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	dst := testStruct{}
	src := testStruct{
		Field1: "some string2",
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructInterfaceSimpleSameType(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	dst := testStruct{
		Field1: "another string",
	}
	src := testStruct{
		Field1: "some string",
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructInterfaceSimpleDifferentType(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	dst := testStruct{
		Field1: 123,
	}
	src := testStruct{
		Field1: "some string",
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructInterfaceStruct(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	type innerStruct struct {
		Field1 int
	}

	dst := testStruct{}
	src := testStruct{
		Field1: innerStruct{
			Field1: 3,
		},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructInterfaceStructSameType(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	type innerStruct struct {
		Field1 int
	}

	dst := testStruct{
		Field1: innerStruct{
			Field1: 4,
		},
	}
	src := testStruct{
		Field1: innerStruct{
			Field1: 3,
		},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructInterfaceStructDifferentType(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	type innerStruct struct {
		Field1 int
	}
	type innerStruct2 struct {
		Field1 string
	}

	dst := testStruct{
		Field1: innerStruct2{
			Field1: "some string",
		},
	}
	src := testStruct{
		Field1: innerStruct{
			Field1: 3,
		},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructInterfaceSlice(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	dst := testStruct{}
	src := testStruct{
		Field1: []int{1, 2, 3},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
}

func TestProtectingCopyStructInterfaceSliceNil1(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	var v []int

	dst := testStruct{
		Field1: v,
	}
	src := testStruct{
		Field1: []int{1, 2, 3},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructInterfaceSliceNil2(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	var v []int

	dst := testStruct{}
	src := testStruct{
		Field1: v,
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructInterfaceSliceOverwrite(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	v := make([]int, 3, 3)
	dst := testStruct{
		Field1: v,
	}
	src := testStruct{
		Field1: []int{1, 2, 3},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectSame(t, v, dst.Field1)
}

func TestProtectingCopyStructInterfaceSliceReplaceExtend(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	v := make([]int, 0, 3)
	dst := testStruct{
		Field1: v,
	}
	src := testStruct{
		Field1: []int{1, 2, 3},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectNotSame(t, v, dst.Field1)
}

func TestProtectingCopyStructInterfaceSliceReplaceShrink(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	v := []int{1, 2, 3, 4, 5}
	dst := testStruct{
		Field1: v,
	}
	src := testStruct{
		Field1: []int{1, 2, 3},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectNotSame(t, v, dst.Field1)
}

func TestProtectingCopyStructInterfaceSliceReplace(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	dst := testStruct{
		Field1: []string{"1", "2", "3"},
	}
	src := testStruct{
		Field1: []int{1, 2, 3},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
}

func TestProtectingCopyStructInterfaceMap(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	dst := testStruct{}
	src := testStruct{
		Field1: map[string]int{"key1": 1, "key2": 2, "key3": 3},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
}

func TestProtectingCopyStructInterfaceMapOverwrite(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	v := map[string]int{"key1": 2, "key2": 2, "key4": 3}
	dst := testStruct{
		Field1: v,
	}
	src := testStruct{
		Field1: map[string]int{"key1": 1, "key2": 2, "key3": 3},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectSame(t, v, dst.Field1)
}

func TestProtectingCopyStructInterfaceMapReplace(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	v := map[string]string{"key1": "2", "key2": "2", "key4": "3"}
	dst := testStruct{
		Field1: v,
	}
	src := testStruct{
		Field1: map[string]int{"key1": 1, "key2": 2, "key3": 3},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectNotSame(t, v, dst.Field1)
}

func TestProtectingCopyStructInterfacePtr(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	dst := testStruct{}
	v := "some string"
	src := testStruct{
		Field1: &v,
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
}

func TestProtectingCopyStructInterfacePtrReplace(t *testing.T) {
	type testStruct struct {
		Field1 interface{}
	}

	v1 := 1
	dst := testStruct{
		Field1: &v1,
	}
	v2 := "some string"
	src := testStruct{
		Field1: &v2,
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectNotSame(t, &v1, dst.Field1)
}

func TestProtectingCopyStructStruct(t *testing.T) {
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
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructAnonymousStruct(t *testing.T) {
	type testStruct struct {
		Field1 struct{ Value int }
	}

	dst := testStruct{}
	src := testStruct{
		Field1: struct{ Value int }{Value: 1234567},
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
}

func TestProtectingCopyStructPointerToSimple(t *testing.T) {
	type testStruct struct {
		Field1 *int
	}

	dst := testStruct{}
	v := 1
	src := testStruct{
		Field1: &v,
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
}

func TestProtectingCopyStructPointerToSimpleOverwrite(t *testing.T) {
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
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectSame(t, &v1, dst.Field1)
}

func TestProtectingCopyStructPointerToStruct(t *testing.T) {
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
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
}

func TestProtectingCopyStructPointerToStructOverwrite(t *testing.T) {
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
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectSame(t, &v1, dst.Field1)
}

func TestProtectingCopyStructPointerToSlice(t *testing.T) {
	type testStruct struct {
		Field1 *[]int
	}

	dst := testStruct{}
	v := []int{1, 2, 3}
	src := testStruct{
		Field1: &v,
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectNotSame(t, *src.Field1, *dst.Field1)
}

func TestProtectingCopyStructPointerToSliceOverwrite(t *testing.T) {
	type testStruct struct {
		Field1 *[]int
	}

	v1 := []int{1, 2, 3}
	v1orig := v1
	v2 := []int{4, 5, 6}
	dst := testStruct{
		Field1: &v1,
	}
	src := testStruct{
		Field1: &v2,
	}
	expectSame(t, v1orig, v1)
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectSame(t, v1orig, v1)
}

func TestProtectingCopyStructPointerToSliceReplaceShirink(t *testing.T) {
	type testStruct struct {
		Field1 *[]int
	}

	v1 := []int{1, 2, 3, 4}
	v1orig := v1
	v2 := []int{4, 5, 6}
	dst := testStruct{
		Field1: &v1,
	}
	src := testStruct{
		Field1: &v2,
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectSliceNotInSameAddr(t, v1orig, v1)
}

func TestProtectingCopyStructPointerToMap(t *testing.T) {
	type testStruct struct {
		Field1 *map[string]int
	}

	dst := testStruct{}
	v := map[string]int{"key1": 1, "key2": 2, "key3": 3, }
	src := testStruct{
		Field1: &v,
	}
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectNotSame(t, src.Field1, dst.Field1)
	expectNotSame(t, *src.Field1, *dst.Field1)
}

func TestProtectingCopyStructPointerToMapOverwrite(t *testing.T) {
	type testStruct struct {
		Field1 *map[string]int
	}

	v1 := map[string]int{"key1": 1, "key2": 102, "key4": 4, }
	v1orig := v1
	v2 := map[string]int{"key1": 1, "key2": 2, "key3": 3, }
	dst := testStruct{
		Field1: &v1,
	}
	src := testStruct{
		Field1: &v2,
	}
	expectSame(t, v1orig, v1)
	expectEquals(
		t,
		nil,
		ProtectingCopy(&dst, &src, ""),
	)
	expectEquals(t, src, dst)
	expectSame(t, v1orig, v1)
}


type exampleUser struct {
	Name string `json:"name"`
	Password string `json:"password" protectfor:"update"`
}

func TestProtectingBindStruct(t* testing.T) {
	binder := func(dst interface{}) error {
		return json.Unmarshal(
			[]byte(
				`{` +
					`"name": "newname",` +
					`"password": "sesame"` +
					`}`,
			),
			dst,
		);
	}
	user := exampleUser{
		Name: "oldname",
		Password: "secret",
	}
	if err := ProtectingBind(binder, &user, "update"); err != nil {
		t.Fatalf("Expect no err, but: %v", err)
	}
	expectEquals(
		t,
		exampleUser{
			Name: "newname",
			Password: "secret",
		},
		user,
	)
}

func TestProtectingBindSlice(t* testing.T) {
	binder := func(dst interface{}) error {
		return json.Unmarshal(
			[]byte(
				`[{` +
					`"name": "newname1",` +
					`"password": "sesame1"` +
					`},` +
					`{` +
					`"name": "newname2",` +
					`"password": "sesame2"` +
					`}]`,
			),
			dst,
		);
	}
	users := []exampleUser{}
	if err := ProtectingBind(binder, &users, "update"); err != nil {
		t.Fatalf("Expect no err, but: %v", err)
	}
	expectEquals(
		t,
		[]exampleUser {
			exampleUser{
				Name: "newname1",
				Password: "",
			},
			exampleUser{
				Name: "newname2",
				Password: "",
			},
		},
		users,
	)
}

func TestProtectingBindMap(t* testing.T) {
	binder := func(dst interface{}) error {
		return json.Unmarshal(
			[]byte(
				`{` +
					`"user1": {` +
					`  "name": "newname1",` +
					`  "password": "sesame1"` +
					`},` +
					`"user2": {` +
					`  "name": "newname2",` +
					`  "password": "sesame2"` +
					`}}`,
			),
			dst,
		);
	}
	users := map[string]*exampleUser{
			"user1": &exampleUser{
				Name: "oldname1",
				Password: "password1",
			},
			"user2": &exampleUser{
				Name: "oldname2",
				Password: "password2",
			},
			"user3": &exampleUser{
				Name: "oldname3",
				Password: "password3",
			},
	}
	if err := ProtectingBind(binder, &users, "update"); err != nil {
		t.Fatalf("Expect no err, but: %v", err)
	}
	expectEquals(
		t,
		map[string]*exampleUser {
			"user1": &exampleUser{
				Name: "newname1",
				Password: "password1",
			},
			"user2": &exampleUser{
				Name: "newname2",
				Password: "password2",
			},
		},
		users,
	)
}

func TestProtectingBindNonPtr(t* testing.T) {
	binder := func(dst interface{}) error {
		return nil
	}
	users := []exampleUser{}
	if err := ProtectingBind(binder, users, "update"); err == nil {
		t.Fatalf("Expect err, but nil: %v", err)
	} else if _, ok := err.(*ErrCopyValueInvalid); !ok {
		t.Fatalf("Expect ErrCopyValueInvalid, but was: %#v", err)
	}
}

func TestProtectingBindError(t* testing.T) {
	expectErr := errors.New("Some error")
	binder := func(dst interface{}) error {
		return expectErr
	}
	var user exampleUser
	err := ProtectingBind(binder, &user, "update")
	expectSame(t, expectErr, err)
}

func TestProtectingErrCopyValueInvalid(t* testing.T) {
	msg := "test message"
	err := &ErrCopyValueInvalid{msg: msg}
	expectEquals(t, msg, err.Error())
}
