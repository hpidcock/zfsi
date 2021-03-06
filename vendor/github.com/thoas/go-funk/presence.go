package funk

import (
	"reflect"
	"strings"
)

// Filter iterates over elements of collection, returning an array of
// all elements predicate returns truthy for.
func Filter(arr interface{}, predicate interface{}) interface{} {
	if !IsIteratee(arr) {
		panic("First parameter must be an iteratee")
	}

	if !IsFunction(predicate, 1, 1) {
		panic("Second argument must be function")
	}

	funcValue := reflect.ValueOf(predicate)

	funcType := funcValue.Type()

	if funcType.Out(0).Kind() != reflect.Bool {
		panic("Return argument should be a boolean")
	}

	arrValue := reflect.ValueOf(arr)

	arrType := arrValue.Type()

	// Get slice type corresponding to array type
	resultSliceType := reflect.SliceOf(arrType.Elem())

	// MakeSlice takes a slice kind type, and makes a slice.
	resultSlice := reflect.MakeSlice(resultSliceType, 0, 0)

	for i := 0; i < arrValue.Len(); i++ {
		elem := arrValue.Index(i)

		result := funcValue.Call([]reflect.Value{elem})[0].Interface().(bool)

		if result {
			resultSlice = reflect.Append(resultSlice, elem)
		}
	}

	return resultSlice.Interface()
}

// Find iterates over elements of collection, returning the first
// element predicate returns truthy for.
func Find(arr interface{}, predicate interface{}) interface{} {
	if !IsIteratee(arr) {
		panic("First parameter must be an iteratee")
	}

	if !IsFunction(predicate, 1, 1) {
		panic("Second argument must be function")
	}

	funcValue := reflect.ValueOf(predicate)

	funcType := funcValue.Type()

	if funcType.Out(0).Kind() != reflect.Bool {
		panic("Return argument should be a boolean")
	}

	arrValue := reflect.ValueOf(arr)

	for i := 0; i < arrValue.Len(); i++ {
		elem := arrValue.Index(i)

		result := funcValue.Call([]reflect.Value{elem})[0].Interface().(bool)

		if result {
			return elem.Interface()
		}
	}

	return nil
}

// IndexOf gets the index at which the first occurrence of value is found in array or return -1
// if the value cannot be found
func IndexOf(in interface{}, elem interface{}) int {
	inValue := reflect.ValueOf(in)

	elemValue := reflect.ValueOf(elem)

	inType := inValue.Type()

	if inType.Kind() == reflect.String {
		return strings.Index(inValue.String(), elemValue.String())
	}

	if inType.Kind() == reflect.Slice {
		for i := 0; i < inValue.Len(); i++ {
			if equal(inValue.Index(i).Interface(), elem) {
				return i
			}
		}
	}

	return -1
}

// LastIndexOf gets the index at which the last occurrence of value is found in array or return -1
// if the value cannot be found
func LastIndexOf(in interface{}, elem interface{}) int {
	inValue := reflect.ValueOf(in)

	elemValue := reflect.ValueOf(elem)

	inType := inValue.Type()

	if inType.Kind() == reflect.String {
		return strings.LastIndex(inValue.String(), elemValue.String())
	}

	if inType.Kind() == reflect.Slice {
		length := inValue.Len()

		for i := length - 1; i >= 0; i-- {
			if equal(inValue.Index(i).Interface(), elem) {
				return i
			}
		}
	}

	return -1
}

// Contains returns true if an element is present in a iteratee.
func Contains(in interface{}, elem interface{}) bool {
	inValue := reflect.ValueOf(in)

	elemValue := reflect.ValueOf(elem)

	inType := inValue.Type()

	if inType.Kind() == reflect.String {
		return strings.Contains(inValue.String(), elemValue.String())
	}

	if inType.Kind() == reflect.Map {
		keys := inValue.MapKeys()
		for i := 0; i < len(keys); i++ {
			if equal(keys[i].Interface(), elem) {
				return true
			}
		}
	}

	if inType.Kind() == reflect.Slice {
		for i := 0; i < inValue.Len(); i++ {
			if equal(inValue.Index(i).Interface(), elem) {
				return true
			}
		}
	}

	return false
}
