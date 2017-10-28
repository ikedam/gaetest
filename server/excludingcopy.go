package server

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	// ExcludingCopyDefaultStructTag is the default tag name to test excluding to copy
	ExcludingCopyDefaultStructTag = "excludingcopy"
)

// ExcludingCopy performs deepcopy excluding fields with the tag
// whose key is "excludingcopy" and value is specified with `toExclude`.
// dst and src must suffice one of followings:
// * src and dst are non-nil pointers of the same type.
// * src and dst are the non-nil map.
// * src and dst are the non-nil slice with the same length.
func ExcludingCopy(dst, src interface{}, toExclude string) error {
	c := &ExcludingCopier{
		ToExclude: toExclude,
	}
	return c.Copy(dst, src)
}

// ErrCopyTypeMismatch represents an error caused for
// type mismatching between source and destination.
type ErrCopyTypeMismatch struct {
	msg string
}

func (e *ErrCopyTypeMismatch) Error() string {
	return e.msg
}

// NewErrCopyTypeMismatch creates a new ErrCopyTypeMismatch.
func NewErrCopyTypeMismatch(dType, sType reflect.Type) *ErrCopyTypeMismatch {
	return &ErrCopyTypeMismatch{
		msg: fmt.Sprintf("types of the destination and the source differ: %v vs %v", dType, sType),
	}
}

// ErrCopyValueInvalid represents a failure of copy
// caused for the source or the destination is invalid
// (e.g. is nil)
type ErrCopyValueInvalid struct {
	msg string
}

func (e *ErrCopyValueInvalid) Error() string {
	return e.msg
}

// ExcludingCopier is the configuration to perform excluding copy
type ExcludingCopier struct {
	// StructTag is the tag name to test fields to exclude to copy.
	// If not specified, ExcludingCopyDefaultStructTag is used.
	StructTag string
	// ToExclude is the tag value to test fields to exclude to copy
	ToExclude string
}

// Copy performs deepcopy excluding fields specified with tag.
// dst and src must suffice one of followings:
// * src and dst are non-nil pointers of the same type.
// * src and dst are the non-nil map.
// * src and dst are the non-nil slice with the same length.
func (c *ExcludingCopier) Copy(dst, src interface{}) error {
	dValue := reflect.ValueOf(dst)
	sValue := reflect.ValueOf(src)
	dType := dValue.Type()
	sType := sValue.Type()

	if !dValue.IsValid() {
		// dst == nil
		return &ErrCopyValueInvalid{
			msg: "Cannot copy as dst is nil",
		}
	}

	if !sValue.IsValid() {
		// src == nil
		return &ErrCopyValueInvalid{
			msg: "Cannot copy as src is nil",
		}
	}

	if dType != sType {
		return NewErrCopyTypeMismatch(dType, sType)
	}

	switch sType.Kind() {
	case reflect.Ptr:
		return c.copyImpl(dValue.Elem(), sValue.Elem())
	case reflect.Map:
		return c.copyMapImpl(dValue, sValue)
	case reflect.Slice:
		if sValue.Len() != dValue.Len() {
			return &ErrCopyValueInvalid{
				msg: "lengths of src and dst must be the same",
			}
		}
		return c.copySliceOrArrayImpl(dValue.Elem(), sValue.Elem())
	}

	return &ErrCopyValueInvalid{
		msg: "dst must be a pointer or a map",
	}
}

// copyImpl is a sub function of ExcludingCopier.Copy
// This assumes values are:
// * CanSet()
func (c *ExcludingCopier) copyImpl(dst, src reflect.Value) error {
	if dst.Type() != src.Type() {
		// This occurs when the type is interface{}
		dst.Set(reflect.Zero(src.Type()))
	}
	switch src.Kind() {
	case reflect.Struct:
		return c.copyStruct(dst, src)
	case reflect.Slice:
		return c.copySlice(dst, src)
	case reflect.Array:
		return c.copySliceOrArrayImpl(dst, src)
	case reflect.Map:
		return c.copyMap(dst, src)
	case reflect.Ptr:
		return c.copyPtr(dst, src)
	}

	// non-structured types
	dst.Set(src)
	return nil
}

// copyStruct is a sub function of ExcludingCopier.Copy
// This assumes values are:
// * reflect.Struct
// * CanSet()
// * types are same
func (c *ExcludingCopier) copyStruct(dst, src reflect.Value) error {
	sType := src.Type()
	tagName := c.StructTag
	if tagName == "" {
		tagName = ExcludingCopyDefaultStructTag
	}
FIELDS:
	for idx := 0; idx < src.NumField(); idx++ {
		tagValues := sType.Field(idx).Tag.Get(tagName)
		if tagValues != "" {
			for _, tagValue := range strings.Split(tagValues, ",") {
				if c.ToExclude == tagValue {
					continue FIELDS
				}
			}
		}
		sValue := src.Field(idx)
		dValue := dst.Field(idx)
		if err := c.copyImpl(dValue, sValue); err != nil {
			return err
		}
	}
	return nil
}

// copyPtr is a sub function of ExcludingCopier.Copy
// This assumes values are:
// * reflect.Ptr
// * CanSet()
// * types are same
func (c *ExcludingCopier) copyPtr(dst, src reflect.Value) error {
	if src.IsNil() {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	if dst.IsNil() {
		dst.Set(reflect.New(src.Type().Elem()))
	} else if dst.Type().Elem() != src.Type().Elem() {
		// This occurs when the type is *interface{}
		// branches to ensure that the test covers this path.
		dst.Set(reflect.New(src.Type().Elem()))
	}
	return c.copyImpl(dst.Elem(), src.Elem())
}

// copySlice is a sub function of ExcludingCopier.Copy
// This assumes values are:
// * reflect.Slice
// * CanSet()
// * types are same
func (c *ExcludingCopier) copySlice(dst, src reflect.Value) error {
	if src.IsNil() {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	if dst.IsNil() || dst.Cap() < src.Len() {
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Len()))
	} else if src.Len() != dst.Len() {
		dst.SetLen(src.Len())
	}

	return c.copySliceOrArrayImpl(dst, src)
}

// copySliceOrArrayImpl is a sub function of ExcludingCopier.Copy
// This assumes values are:
// * reflect.Slice or reflect.Array
// * IsValid()
// * have the same length
// * types are same
func (c *ExcludingCopier) copySliceOrArrayImpl(dst, src reflect.Value) error {
	for idx := 0; idx < src.Len(); idx++ {
		sValue := src.Index(idx)
		dValue := dst.Index(idx)
		if err := c.copyImpl(dValue, sValue); err != nil {
			return err
		}
	}

	return nil
}

// copyMap is a sub function of ExcludingCopier.Copy
// This assumes values are:
// * reflect.Map
// * CanSet()
// * types are same
func (c *ExcludingCopier) copyMap(dst, src reflect.Value) error {
	if src.IsNil() {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	if dst.IsNil() {
		// dst.Set(reflect.MakeMapWithSize(src.Type(), src.Len()))
		dst.Set(reflect.MakeMap(src.Type()))
	}
	return c.copyMapImpl(dst, src)
}

// copyMapImpl is a sub function of ExcludingCopier.Copy
// This assumes values are:
// * reflect.Map
// * types are same
func (c *ExcludingCopier) copyMapImpl(dst, src reflect.Value) error {
	// Remove unnecessary Keys
	for _, key := range dst.MapKeys() {
		if src.MapIndex(key).IsValid() {
			continue
		}
		dst.SetMapIndex(key, reflect.ValueOf(nil))
	}

	for _, key := range src.MapKeys() {
		sValue := src.MapIndex(key)
		dValue := dst.MapIndex(key)

		if c.canSet(dValue, sValue) {
			// pointer type, slice, map, and so on.
			if err := c.copyImpl(dValue, sValue); err != nil {
				return err
			}
		} else {
			dValue = c.createDest(sValue)
			if err := c.copyImpl(dValue, sValue); err != nil {
				return err
			}
			dst.SetMapIndex(key, dValue)
		}
	}
	return nil
}

// canSet tests whether copying src to dest affects the caller.
func (c *ExcludingCopier) canSet(dst, src reflect.Value) bool {
	if !src.IsValid() || !dst.IsValid() {
		return false
	}

	if src.Type() != dst.Type() {
		// this occurs when type is interface{}
		return false
	}

	switch dst.Kind() {
	case reflect.Ptr, reflect.Map:
		return true
	case reflect.Slice:
		return src.Len() == dst.Len()
	}
	return false
}

// createDest creates a new object to perform excludingCopy
func (c *ExcludingCopier) createDest(src reflect.Value) reflect.Value {
	switch src.Kind() {
	case reflect.Slice:
		return reflect.MakeSlice(src.Type(), src.Len(), src.Len())
	case reflect.Map:
		// return reflect.MakeMapWithSize(src.Type(), src.Len())
		return reflect.MakeMap(src.Type())
	case reflect.Ptr:
		return reflect.New(src.Type().Elem())
	}
	return reflect.Zero(src.Type())
}
