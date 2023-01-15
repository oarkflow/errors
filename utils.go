package errors

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

// Code returns the code of the root error, if available.
// Otherwise, returns INTERNAL.
func Code(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.Code != "" {
		return e.Code
	} else if ok && e.Err != nil {
		return Code(e.Err)
	}
	return INTERNAL
}

// Message returns the human-readable message of the error,
// if available. Otherwise, returns a generic error
// message.
func Message(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.Message != "" {
		return e.Message
	} else if ok && e.Err != nil {
		return Message(e.Err)
	}
	return GlobalError
}

// ToError Returns an application error from input. If The type
// is not of type Error, nil will be returned.
func ToError(err any) *Error {
	switch v := err.(type) {
	case *Error:
		return v
	case Error:
		return &v
	case error:
		return &Error{Err: fmt.Errorf(v.Error())}
	case string:
		return &Error{Err: fmt.Errorf(v)}
	default:
		return nil
	}
}

func getLocation() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", file, line)
}

// FromByte converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func FromByte(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// ToByte converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func ToByte(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

func getCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return fmt.Sprintf("%s", runtime.FuncForPC(pc).Name())
}
