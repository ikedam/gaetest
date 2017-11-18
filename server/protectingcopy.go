package server

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	// ProtectingCopyDefaultStructTag is the default tag name to test fields not to copy
	ProtectingCopyDefaultStructTag = "protectfor"
)

// ErrCopyValueInvalid represents a failure of copy
// caused for the source or the destination is invalid
// (e.g. is nil)
type ErrCopyValueInvalid struct {
	msg string
}

func (e *ErrCopyValueInvalid) Error() string {
	return e.msg
}

// ProtectingCopy performs deepcopy excluding fields with the tag
// whose key is "protectfor" and value is specified with `protectFor`.
// dst and src must suffice one of followings:
// * src and dst are non-nil pointers of the same type.
// * src and dst are the non-nil map.
// * src and dst are the non-nil slice with the same length.
func ProtectingCopy(dst, src interface{}, protectFor string) error {
	c := &ProtectingCopier{
		ProtectFor: protectFor,
	}
	return c.Copy(dst, src)
}

// ProtectingBind wraps binging functions such as echo.Context.Bind(),
// providing field protection.
func ProtectingBind(binder func(dst interface{}) error, dst interface{}, protectFor string) error {
	dType := reflect.TypeOf(dst)
	if dType.Kind() != reflect.Ptr {
		return &ErrCopyValueInvalid{
			msg: "dst must be a pointer",
		}
	}
	sValue := reflect.New(dType.Elem())
	if err := binder(sValue.Interface()); err != nil {
		return err
	}
	return ProtectingCopy(dst, sValue.Interface(), protectFor)
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

// ProtectingCopier is the configuration to perform protecting copy
type ProtectingCopier struct {
	// StructTag is the tag name to test fields not to copy.
	// If not specified, ProtectingCopyDefaultStructTag is used.
	StructTag string
	// ProtectFor is the tag value to test fields not to copy
	ProtectFor string
}

// Copy performs deepcopy protecting fields specified with tag.
// dst and src must suffice one of followings:
// * src and dst are non-nil pointers of the same type.
// * src and dst are the non-nil map.
// * src and dst are the non-nil slice with the same length.
func (c *ProtectingCopier) Copy(dst, src interface{}) error {
	dValue := reflect.ValueOf(dst)
	sValue := reflect.ValueOf(src)
	dType := reflect.TypeOf(dst)
	sType := reflect.TypeOf(src)

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
		return c.copySliceOrArrayImpl(dValue, sValue)
	}

	return &ErrCopyValueInvalid{
		msg: "dst must be a pointer or a map",
	}
}

// copyImpl is a sub function of ProtectingCopier.Copy
// This assumes values are:
// * CanSet()
// * Same static types
func (c *ProtectingCopier) copyImpl(dst, src reflect.Value) error {
	switch src.Kind() {
	case reflect.Interface:
		return c.copyInterface(dst, src)
	case reflect.Struct:
		return c.copyStruct(dst, src)
	case reflect.Slice:
		return c.copySlice(dst, src)
	case reflect.Array:
		return c.copyArray(dst, src)
	case reflect.Map:
		return c.copyMap(dst, src)
	case reflect.Ptr:
		return c.copyPtr(dst, src)
	}

	// non-structured types
	dst.Set(src)
	return nil
}

// copyInterface is a sub function of ProtectingCopier.Copy
// This assumes values are:
// * reflect.Interface
// * CanSet()
func (c *ProtectingCopier) copyInterface(dst, src reflect.Value) error {
	if src.IsNil() {
		dst.Set(src)
		return nil
	}
	sType := reflect.TypeOf(src.Interface())
	dType := reflect.TypeOf(dst.Interface())
	switch sType.Kind() {
	case reflect.Slice:
		if src.Elem().IsNil() {
			dst.Set(src)
			return nil
		} else if dType != sType || dst.IsNil() || dst.Elem().IsNil() || dst.Elem().Len() != src.Elem().Len() {
			if _dst, err := c.createCopiedDest(src.Elem()); err == nil {
				dst.Set(_dst)
				return nil
			} else {
				return err
			}
		}
		return c.copyImpl(dst.Elem(), src.Elem())
	case reflect.Map, reflect.Ptr:
		if src.Elem().IsNil() {
			dst.Set(src)
			return nil
		} else if dType != sType || dst.IsNil() || dst.Elem().IsNil() {
			if _dst, err := c.createCopiedDest(src.Elem()); err == nil {
				dst.Set(_dst)
				return nil
			} else {
				return err
			}
		}
		return c.copyImpl(dst.Elem(), src.Elem())
	}

	// non-pointer types in interface are unmodifiable
	if _dst, err := c.createCopiedDest(src.Elem()); err == nil {
		dst.Set(_dst)
	} else {
		return err
	}
	return nil
}

// copyStruct is a sub function of ProtectingCopier.Copy
// This assumes values are:
// * reflect.Struct
// * CanSet()
// * types are same
func (c *ProtectingCopier) copyStruct(dst, src reflect.Value) error {
	sType := reflect.TypeOf(src.Interface())
	tagName := c.StructTag
	if tagName == "" {
		tagName = ProtectingCopyDefaultStructTag
	}
FIELDS:
	for idx := 0; idx < sType.NumField(); idx++ {
		field := sType.Field(idx)
		if field.PkgPath != "" {
			// unexported field. skip.
			continue
		}
		tagValues := field.Tag.Get(tagName)
		if tagValues != "" {
			for _, tagValue := range strings.Split(tagValues, ",") {
				if c.ProtectFor == tagValue {
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

// copyPtr is a sub function of ProtectingCopier.Copy
// This assumes values are:
// * reflect.Ptr
// * CanSet()
// * types are same
func (c *ProtectingCopier) copyPtr(dst, src reflect.Value) error {
	if src.IsNil() {
		dst.Set(src)
		return nil
	}
	if dst.IsNil() {
		dst.Set(reflect.New(reflect.TypeOf(src.Interface()).Elem()))
	}
	return c.copyImpl(dst.Elem(), src.Elem())
}

// copySlice is a sub function of ProtectingCopier.Copy
// This assumes values are:
// * reflect.Slice
// * CanSet()
// * types are same
func (c *ProtectingCopier) copySlice(dst, src reflect.Value) error {
	if src.IsNil() {
		dst.Set(src)
		return nil
	}

	// if dst.Cap() < src.Len() {
	// 	if _dst, err := c.createCopiedDest(src); err == nil {
	// 		dst.Set(_dst)
	// 		return nil
	// 	} else {
	// 		return err
	// 	}
	// } else if src.Len() != dst.Len() {
	// 	dst.SetLen(src.Len())
	// }

	// Safer way.
	if dst.Len() != src.Len() {
		if _dst, err := c.createCopiedDest(src); err == nil {
			dst.Set(_dst)
			return nil
		} else {
			return err
		}
	}

	return c.copySliceOrArrayImpl(dst, src)
}

// copySlice is a sub function of ProtectingCopier.Copy
// This assumes values are:
// * reflect.Array
// * CanSet()
// * types are same
func (c *ProtectingCopier) copyArray(dst, src reflect.Value) error {
	return c.copySliceOrArrayImpl(dst, src)
}

// copySliceOrArrayImpl is a sub function of ProtectingCopier.Copy
// This assumes values are:
// * reflect.Slice or reflect.Array
// * IsValid()
// * have the same length
// * types are same
func (c *ProtectingCopier) copySliceOrArrayImpl(dst, src reflect.Value) error {
	for idx := 0; idx < src.Len(); idx++ {
		sValue := src.Index(idx)
		dValue := dst.Index(idx)
		if err := c.copyImpl(dValue, sValue); err != nil {
			return err
		}
	}

	return nil
}

// copyMap is a sub function of ProtectingCopier.Copy
// This assumes values are:
// * reflect.Map
// * CanSet()
// * types are same
func (c *ProtectingCopier) copyMap(dst, src reflect.Value) error {
	if src.IsNil() {
		dst.Set(src)
		return nil
	}
	if dst.IsNil() {
		// dst.Set(reflect.MakeMapWithSize(reflect.TypeOf(src.Interface()), src.Len()))
		dst.Set(reflect.MakeMap(reflect.TypeOf(src.Interface())))
	}
	return c.copyMapImpl(dst, src)
}

// copyMapImpl is a sub function of ProtectingCopier.Copy
// This assumes values are:
// * reflect.Map
// * types are same
func (c *ProtectingCopier) copyMapImpl(dst, src reflect.Value) error {
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

		if c.canSetForMap(dValue, sValue) {
			// pointer type, slice, map, and so on.
			if err := c.copyImpl(dValue, sValue); err != nil {
				return err
			}
		} else {
			if _dValue, err := c.createCopiedDest(sValue); err == nil {
				dst.SetMapIndex(key, _dValue)
			} else {
				return err
			}
		}
	}
	return nil
}

// canSetForMap tests whether copying src to dst affects the caller.
func (c *ProtectingCopier) canSetForMap(dst, src reflect.Value) bool {
	if !src.IsValid() || !dst.IsValid() {
		return false
	}

	dType := reflect.TypeOf(dst.Interface())
	sType := reflect.TypeOf(src.Interface())

	if sType != dType {
		// this occurs when type is interface{}
		return false
	}

	switch dType.Kind() {
	case reflect.Ptr, reflect.Map:
		for src.Kind() == reflect.Interface && src.IsValid() {
			src = src.Elem()
		}
		for dst.Kind() == reflect.Interface && dst.IsValid() {
			dst = dst.Elem()
		}
		return src.IsValid() && !src.IsNil() && dst.IsValid() && !dst.IsNil()
	case reflect.Slice:
		for src.Kind() == reflect.Interface && src.IsValid() {
			src = src.Elem()
		}
		for dst.Kind() == reflect.Interface && dst.IsValid() {
			dst = dst.Elem()
		}
		return src.IsValid() && !src.IsNil() && dst.IsValid() && !dst.IsNil() && src.Len() == dst.Len()
	}
	return dst.CanSet()
}

// createCopiedDest creates a new object and perform protecting copy
func (c *ProtectingCopier) createCopiedDest(src reflect.Value) (reflect.Value, error) {
	origSrc := src
	for src.Kind() == reflect.Interface && src.IsValid() {
		src = src.Elem()
	}
	if !src.IsValid() {
		return origSrc, nil
	}
	switch src.Kind() {
	case reflect.Slice:
		if src.IsNil() {
			return origSrc, nil
		}
		dst := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
		err := c.copyImpl(dst, src)
		return dst, err
	case reflect.Map:
		if src.IsNil() {
			return origSrc, nil
		}
		// dst := reflect.MakeMapWithSize(src.Type(), src.Len())
		dst := reflect.MakeMap(src.Type())
		err := c.copyImpl(dst, src)
		return dst, err
	case reflect.Ptr:
		if src.IsNil() {
			return origSrc, nil
		}
		dst := reflect.New(src.Type().Elem())
		err := c.copyImpl(dst, src)
		return dst, err
	}
	dst :=reflect.New(src.Type()).Elem()
	err := c.copyImpl(dst, src)
	return dst, err
}
