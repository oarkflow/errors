package errors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

type Trace struct {
	Index    int    `json:"index"`
	Function string `json:"function,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

func (t Trace) String() string {
	return fmt.Sprintf("#%d %s:%d %s", t.Index, t.File, t.Line, t.Function)
}

type StackTrace []Trace

func (t StackTrace) String() string {
	var buf bytes.Buffer
	for _, trace := range t {
		buf.WriteString(trace.String())
		buf.WriteString("\n")
	}
	return buf.String()
}

func (t StackTrace) StringArray() (s []string) {
	for _, trace := range t {
		s = append(s, trace.String())
	}
	return
}

// New is a wrapper for the stdlib new function.
func New(message string) error {
	return errors.New(message)
}

// Unwrap calls the stdlib errors.UnUnwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Is calls the stdlib errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As calls the stdlib errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// NewInternal returns an Error with a INTERNAL error code.
func NewInternal(err error, message, op string) *Error {
	return newError(err, message, INTERNAL, op)
}

// NewConflict returns an Error with a CONFLICT error code.
func NewConflict(err error, message, op string) *Error {
	return newError(err, message, CONFLICT, op)
}

// NewInvalid returns an Error with a INVALID error code.
func NewInvalid(err error, message, op string) *Error {
	return newError(err, message, INVALID, op)
}

// NewNotFound returns an Error with a NOTFOUND error code.
func NewNotFound(err error, message, op string) *Error {
	return newError(err, message, NOTFOUND, op)
}

// NewUnknown returns an Error with a UNKNOWN error code.
func NewUnknown(err error, message, op string) *Error {
	return newError(err, message, UNKNOWN, op)
}

// NewMaximumAttempts returns an Error with a MAXIMUMATTEMPTS error code.
func NewMaximumAttempts(err error, message, op string) *Error {
	return newError(err, message, MAXIMUMATTEMPTS, op)
}

// NewExpired returns an Error with a EXPIRED error code.
func NewExpired(err error, message, op string) *Error {
	return newError(err, message, EXPIRED, op)
}

// NewE returns an Error with the DefaultCode.
func NewE(err error, message, op string) *Error {
	return newError(err, message, DefaultCode, op)
}

// ErrorF returns an Error with the DefaultCode and
// formatted message arguments.
func ErrorF(err error, op, format string, args ...any) *Error {
	return NewE(err, fmt.Sprintf(format, args...), op)
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message, op string) *Error {
	if err == nil {
		return nil
	}
	return NewE(err, message, op)
}

// newError is an alias for New by creating the pcs
// file line and constructing the error message.
func newError(err error, message, code, op string) *Error {
	_, file, line, _ := runtime.Caller(2)
	pcs := make([]uintptr, 2)
	_ = runtime.Callers(2, pcs)
	var stackTrace StackTrace
	for i, pc := range pcs {
		p := runtime.FuncForPC(pc)
		f, l := p.FileLine(pc)
		stackTrace = append(stackTrace, Trace{
			Index:    i,
			Function: p.Name(),
			File:     f,
			Line:     l,
		})
	}
	e := &Error{
		Code:       code,
		Message:    message,
		Operation:  op,
		Err:        err,
		Additional: stackTrace,
		fileLine:   file + ":" + strconv.Itoa(line),
		pcs:        pcs,
	}
	if code == INTERNAL {
		e.Internal = true
	}
	return e
}

// Application error codes.
const (
	// CONFLICT - An action cannot be performed.
	CONFLICT = "conflict"
	// INTERNAL - Error within the application.
	INTERNAL = "internal"
	// INVALID - Validation failed.
	INVALID = "invalid"
	// NOTFOUND - Entity does not exist.
	NOTFOUND = "not_found"
	// UNKNOWN - Application unknown error.
	UNKNOWN = "unknown"
	// MAXIMUMATTEMPTS - More than allowed action.
	MAXIMUMATTEMPTS = "maximum_attempts"
	// EXPIRED - Subscription expired.
	EXPIRED = "expired"
)

var (
	// DefaultCode is the default code returned when
	// none is specified.
	DefaultCode = INTERNAL
	// GlobalError is a general message when no error message
	// has been found.
	GlobalError = "An error has occurred."
)

// Error defines a standard application error.
type Error struct {
	Code       string     `json:"code"`
	Message    string     `json:"message"`
	Operation  string     `json:"operation"`
	Err        error      `json:"error"`
	Additional StackTrace `json:"additional"`
	Internal   bool       `json:"internal"`
	fileLine   string
	pcs        []uintptr
}

// Error returns the string representation of the error
// message by implementing the error interface.
func (e *Error) Error() string {
	var buf bytes.Buffer

	// Print the error code if there is one.
	if e.Code != "" {
		buf.WriteString("<" + e.Code + "> ")
	}

	// Print the file-line, if any.
	if e.fileLine != "" {
		buf.WriteString(e.fileLine + " - ")
	}

	// Print the current operation in our stack, if any.
	if e.Operation != "" {
		buf.WriteString(e.Operation + ": ")
	}

	// Print the original error message, if any.
	if e.Err != nil {
		buf.WriteString(e.Err.Error() + ", ")
	}

	// Print the message, if any.
	if e.Message != "" {
		buf.WriteString(e.Message)
	}

	return strings.TrimSuffix(strings.TrimSpace(buf.String()), ",")
}

func (e *Error) ErrorWithStackTrace() string {
	var buf bytes.Buffer
	buf.WriteString("Type: ")
	buf.WriteString(e.Code)
	buf.WriteString(", Message: ")
	buf.WriteString(e.Message)
	buf.WriteString(", Operation: ")
	buf.WriteString(e.Operation)
	buf.WriteString("\n")
	buf.WriteString(e.Additional.String())
	switch er := e.Err.(type) {
	case *Error:
		buf.WriteString("\n")
		buf.WriteString(er.ErrorWithStackTrace())
		buf.WriteString("\n")
	case error:
		buf.WriteString("\n")
		buf.WriteString(er.Error())
	}
	return buf.String()
}

// FileLine returns the file and line in which the error
// occurred.
func (e *Error) FileLine() string {
	return e.fileLine
}

// Unwrap unwraps the original error message.
func (e *Error) Unwrap() error {
	return e.Err
}

// HTTPStatusCode is a convenience method used to get the appropriate
// HTTP response status code for the respective error type.
func (e *Error) HTTPStatusCode() int {
	status := http.StatusInternalServerError
	switch e.Code {
	case CONFLICT:
		return http.StatusConflict
	case INVALID:
		return http.StatusBadRequest
	case NOTFOUND:
		return http.StatusNotFound
	case EXPIRED:
		return http.StatusPaymentRequired
	case MAXIMUMATTEMPTS:
		return http.StatusTooManyRequests
	}
	return status
}

// RuntimeFrames returns function/file/line information.
func (e *Error) RuntimeFrames() *runtime.Frames {
	return runtime.CallersFrames(e.pcs)
}

// ProgramCounters returns the slice of PC values associated
// with the error.
func (e *Error) ProgramCounters() []uintptr {
	return e.pcs
}

// StackTrace returns a string representation of the errors
// stacktrace, where each trace is separated by a newline
// and tab '\t'.
func (e *Error) StackTrace() string {
	trace := make([]string, 0, 100)
	rFrames := e.RuntimeFrames()
	frame, ok := rFrames.Next()
	line := strconv.Itoa(frame.Line)
	trace = append(trace, frame.Function+"(): "+e.Message)

	for ok {
		trace = append(trace, "\t"+frame.File+":"+line)
		frame, ok = rFrames.Next()
	}

	return strings.Join(trace, "\n")
}

// StackTraceSlice returns a string slice of the errors
// stacktrace.
func (e *Error) StackTraceSlice() []string {
	trace := make([]string, 0, 100)
	rFrames := e.RuntimeFrames()
	frame, ok := rFrames.Next()
	line := strconv.Itoa(frame.Line)
	trace = append(trace, frame.Function+"(): "+e.Message)

	for ok {
		trace = append(trace, frame.File+":"+line)
		frame, ok = rFrames.Next()
	}

	return trace
}

// wrappingError is the wrapping error features the error
// and file line in strings suitable for json.Marshal.
type wrappingError struct {
	Code       string     `json:"code"`
	Message    string     `json:"message"`
	Operation  string     `json:"operation"`
	Err        string     `json:"error"`
	FileLine   string     `json:"file_line"`
	Additional StackTrace `json:"additional"`
	Internal   bool       `json:"internal"`
}

// MarshalJSON implements encoding/Marshaller to wrap the
// error as a string if there is one.
func (e *Error) MarshalJSON() ([]byte, error) {
	err := wrappingError{
		Code:       e.Code,
		Message:    e.Message,
		Operation:  e.Operation,
		Additional: e.Additional,
		Internal:   e.Internal,
	}
	if e.Err != nil {
		err.Err = e.Err.Error()
		err.FileLine = e.fileLine
	}
	return json.Marshal(err)
}

func (e *Error) JSONAsString() (string, error) {
	bt, err := e.MarshalJSON()
	return FromByte(bt), err
}

// UnmarshalJSON implements encoding/Marshaller to unmarshal
// the wrapping error to type Error.
func (e *Error) UnmarshalJSON(data []byte) error {
	var err wrappingError
	mErr := json.Unmarshal(data, &err)
	if mErr != nil {
		return mErr
	}
	e.Code = err.Code
	e.Message = err.Message
	e.Operation = err.Operation
	e.Additional = err.Additional
	e.Internal = err.Internal
	e.fileLine = err.FileLine
	if err.Err != "" {
		e.Err = errors.New(err.Err)
	}
	return nil
}

// Scan implements the sql.Scanner interface.
func (e *Error) Scan(value any) error {
	if value == nil {
		return nil
	}
	buf, ok := value.([]byte)
	if !ok || buf == nil {
		return fmt.Errorf("scan not supported for *errors.Error")
	}
	return json.Unmarshal(buf, e)
}

// Value implements the driver.Valuer interface.
func (e *Error) Value() ([]byte, error) {
	return json.Marshal(e)
}
